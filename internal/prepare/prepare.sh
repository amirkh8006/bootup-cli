#!/usr/bin/env bash
# Common preparation for Debian/Ubuntu servers using systemd.
set -Eeuo pipefail

LOG_FILE=/var/log/bootup-prepare.log
HOSTNAME_NEW=
TIMEZONE_TARGET=
SWAP_SIZE_MIB=2048
CONFIGURE_FIREWALL=false
SSH_PORT=22
SSH_SOURCE=
UPGRADE_PACKAGES=false

usage() {
    cat <<'HELP'
Usage: sudo ./commonRoles.sh [options]
  --hostname NAME       Set a unique server hostname (default: preserve)
  --timezone ZONE       Set timezone (default: preserve)
  --swap-mib SIZE       Create swap if none exists (default: 2048; 0 disables)
  --firewall            Enable UFW, preserving existing rules
  --no-firewall         Leave firewall settings unchanged (default)
  --ssh-port PORT       Actual SSH port to allow (default: 22)
  --ssh-source CIDR     Optional management IP/CIDR (default: any source)
  --upgrade            Apply all available package upgrades
  --help                Show this help

Default: install common tools and chrony, verify time synchronization, and report
resources. Existing hostnames, timezone, active swap, and firewall rules are preserved.
UFW allows HTTP (80/tcp) and HTTPS (443/tcp), denies other incoming traffic, and
allows outgoing traffic. SSH is allowed from any
source unless --ssh-source is supplied. Existing broader rules remain effective.
For custom SSH ports, pass --ssh-port. Add application rules separately as needed.
HELP
}

parse_args() {
    while (($#)); do
        case "$1" in
            --hostname|--timezone|--swap-mib|--ssh-port|--ssh-source)
                (($# >= 2)) && [[ -n $2 ]] || { echo "Missing value for $1" >&2; exit 2; }
                case "$1" in
                    --hostname) HOSTNAME_NEW=$2 ;;
                    --timezone) TIMEZONE_TARGET=$2 ;;
                    --swap-mib) SWAP_SIZE_MIB=$2 ;;
                    --ssh-port) SSH_PORT=$2 ;;
                    --ssh-source) SSH_SOURCE=$2 ;;
                esac
                shift 2 ;;
            --firewall) CONFIGURE_FIREWALL=true; shift ;;
            --no-firewall) CONFIGURE_FIREWALL=false; shift ;;
            --upgrade) UPGRADE_PACKAGES=true; shift ;;
            --help|-h) usage; exit 0 ;;
            *) echo "Unknown option: $1" >&2; usage >&2; exit 2 ;;
        esac
    done
    [[ $SWAP_SIZE_MIB =~ ^[0-9]{1,7}$ ]] || { echo 'Invalid swap size' >&2; exit 2; }
    SWAP_SIZE_MIB=$((10#$SWAP_SIZE_MIB))
    if [[ -n $HOSTNAME_NEW ]]; then
        [[ ${#HOSTNAME_NEW} -le 63 && $HOSTNAME_NEW =~ ^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?$ ]] || {
            echo 'Hostname must be a single DNS label, at most 63 characters' >&2; exit 2;
        }
    fi
    if $CONFIGURE_FIREWALL; then
        [[ $SSH_PORT =~ ^[0-9]{1,5}$ && $SSH_SOURCE != -* ]] || {
            echo 'Invalid SSH port or source' >&2; exit 2;
        }
        SSH_PORT=$((10#$SSH_PORT))
        ((SSH_PORT >= 1 && SSH_PORT <= 65535)) || { echo 'Invalid SSH port' >&2; exit 2; }
    fi
}

log() { printf '%s - %s\n' "$(date '+%F %T%z')" "$*"; }
die() { log "ERROR: $*" >&2; exit 1; }

preflight() {
    ((EUID == 0)) || { echo 'Run this script as root' >&2; exit 1; }
    [[ -r /etc/os-release ]] || die 'Cannot determine operating system'
    # shellcheck source=/dev/null
    . /etc/os-release
    case "${ID:-}" in debian|ubuntu) ;; *) die 'Only Debian and Ubuntu are supported' ;; esac
    [[ -d /run/systemd/system ]] || die 'A running systemd host is required'
    local command_name
    for command_name in apt-get dpkg-query hostnamectl timedatectl swapon mkswap findmnt flock; do
        command -v "$command_name" >/dev/null || die "Missing prerequisite: $command_name"
    done
    if [[ -n $TIMEZONE_TARGET ]]; then
        [[ $TIMEZONE_TARGET != /* && $TIMEZONE_TARGET != *..* && -f /usr/share/zoneinfo/$TIMEZONE_TARGET ]] || die 'Invalid timezone'
    fi
    exec 9>/run/lock/bootup-prepare.lock
    flock -n 9 || die 'Another preparation process is running'
    touch "$LOG_FILE"
    chmod 600 "$LOG_FILE"
    exec > >(tee -a "$LOG_FILE") 2>&1
    trap 'rc=$?; printf "ERROR: line %s failed (exit %s)\n" "$LINENO" "$rc" >&2; exit "$rc"' ERR
}

install_packages() {
    log 'Installing common tools and chrony (the selected time synchronization service)'
    export DEBIAN_FRONTEND=noninteractive
    apt-get update
    apt-get install -y curl wget htop net-tools dnsutils vim git unzip tar gzip \
        bzip2 zip ca-certificates gnupg lsb-release rsync tree screen tmux chrony
    if $CONFIGURE_FIREWALL; then apt-get install -y ufw; fi
    if $UPGRADE_PACKAGES; then
        log 'Applying all available package upgrades'
        apt-get upgrade --with-new-pkgs -y
    fi
    systemctl enable --now chrony.service
}

configure_identity() {
    if [[ -n $HOSTNAME_NEW ]]; then
        log "Setting hostname to $HOSTNAME_NEW"
        # Preserve all existing aliases; manage only our own additional entry.
        local hosts_tmp
        hosts_tmp=$(mktemp)
        awk '$0 !~ /# common-roles managed hostname$/ {print}' /etc/hosts > "$hosts_tmp"
        printf '127.0.1.1\t%s # common-roles managed hostname\n' "$HOSTNAME_NEW" >> "$hosts_tmp"
        cp -p /etc/hosts "/etc/hosts.backup.$(date +%s%N)"
        cat "$hosts_tmp" > /etc/hosts
        rm -f "$hosts_tmp"
        hostnamectl set-hostname "$HOSTNAME_NEW"
    fi
    if [[ -n $TIMEZONE_TARGET ]]; then
        timedatectl set-timezone "$TIMEZONE_TARGET"
    fi
}

configure_swap() {
    ((SWAP_SIZE_MIB > 0)) || return 0
    if [[ -n $(swapon --show --noheadings) ]]; then
        log 'Active swap exists; preserving it'
        return
    fi
    # Refuse to overwrite an existing file, symlink, or configured inactive swap.
    if [[ -e /swapfile || -L /swapfile ]] || awk '$1 !~ /^#/ && $3 == "swap" {found=1} END {exit !found}' /etc/fstab; then
        log 'WARNING: Inactive swap file or fstab swap entry exists; review it manually'
        return
    fi
    local filesystem
    filesystem=$(findmnt -no FSTYPE -T /)
    case "$filesystem" in
        ext4|xfs) ;;
        *) log "WARNING: Automatic swap creation skipped on $filesystem"; return ;;
    esac
    local available_mib
    available_mib=$(df -Pm / | awk 'END {print $4}')
    if ((available_mib < SWAP_SIZE_MIB + 1024)); then
        log 'WARNING: Insufficient disk space for requested swap plus 1 GiB reserve; skipping'
        return
    fi
    log "Creating ${SWAP_SIZE_MIB} MiB swap file"
    # Exclusive creation prevents overwriting an existing file.
    (umask 077; set -o noclobber; : > /swapfile)
    dd if=/dev/zero of=/swapfile bs=1M count="$SWAP_SIZE_MIB" status=progress conv=fsync
    chmod 600 /swapfile
    mkswap /swapfile
    swapon /swapfile
    printf '/swapfile none swap sw 0 0\n' >> /etc/fstab
}

configure_firewall() {
    $CONFIGURE_FIREWALL || return 0
    if [[ -n $SSH_SOURCE ]]; then
        log "Allowing SSH port $SSH_PORT from $SSH_SOURCE before enabling UFW"
        ufw allow from "$SSH_SOURCE" to any port "$SSH_PORT" proto tcp
    else
        log "Allowing SSH port $SSH_PORT from any source before enabling UFW"
        ufw allow "$SSH_PORT/tcp"
    fi
    ufw allow 80/tcp
    ufw allow 443/tcp
    ufw default deny incoming
    ufw default allow outgoing
    ufw --force enable
    ufw status verbose
}

report() {
    log 'System preparation report'
    hostnamectl
    free -h
    df -hT -x tmpfs -x devtmpfs
    df -i /
    swapon --show
    log "Available CPU cores: $(nproc)"
    if chronyc tracking; then
        if ! chronyc waitsync 6 0.5 0 2; then
            log 'WARNING: Clock is not yet synchronized; check NTP connectivity and sources'
        fi
    else
        log 'WARNING: Unable to query chrony'
    fi
    local ssh_unit
    for ssh_unit in ssh.service sshd.service; do
        if systemctl is-active --quiet "$ssh_unit"; then
            log "SSH is active: $ssh_unit"
            break
        fi
    done
    if ! systemctl is-active --quiet ssh.service && ! systemctl is-active --quiet sshd.service; then
        log 'WARNING: No active SSH service detected'
    fi
    if [[ -f /var/run/reboot-required ]]; then log 'WARNING: A reboot is required'; fi
    log 'Preparation completed. Review warnings and configure application-specific access, backups, and monitoring separately.'
    log "Log: $LOG_FILE"
}

main() {
    parse_args "$@"
    preflight
    log 'Starting common server preparation'
    install_packages
    configure_identity
    configure_swap
    configure_firewall
    report
}

if [[ ${BASH_SOURCE[0]} == "$0" ]]; then main "$@"; fi

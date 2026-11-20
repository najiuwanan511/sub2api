#!/bin/bash
# =============================================================================
# Sub2API Docker Deployment Preparation Script
# =============================================================================
# This script prepares deployment files for Sub2API:
#   - Downloads docker-compose.local.yml and .env.example
#   - Generates secure secrets (JWT_SECRET, TOTP_ENCRYPTION_KEY, POSTGRES_PASSWORD)
#   - Generates a random admin login email (ADMIN_EMAIL)
#   - Creates necessary data directories
#
# After running this script, you can start services with:
#   docker compose up -d
# Use --domain api.example.com to enable automatic HTTPS via Caddy.
# =============================================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# GitHub raw content base URL
GITHUB_RAW_URL="https://raw.githubusercontent.com/najiuwanan511/sub2api/main/deploy"

# Print colored message
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Generate random secret
generate_secret() {
    openssl rand -hex 32
}

# Generate a random admin login email so new installs never use a guessable default
generate_admin_email() {
    local suffix
    suffix=$(openssl rand -hex 6) || return 1
    [ -n "$suffix" ] || return 1
    echo "admin-${suffix}@sub2api.local"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Only accept DNS names, not URLs, ports, wildcards or config syntax.
valid_domain() {
    local domain="$1" label
    local -a labels
    local LC_ALL=C
    [[ ${#domain} -le 253 && "$domain" =~ ^[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)+$ ]] || return 1
    IFS=. read -r -a labels <<< "$domain"
    [[ "${labels[${#labels[@]}-1]}" =~ [a-zA-Z] ]] || return 1
    for label in "${labels[@]}"; do
        [[ ${#label} -le 63 && "$label" != -* && "$label" != *- ]] || return 1
    done
}

download_file() {
    local source="$1" target="$2"
    if command_exists curl; then
        curl -fsSL "${GITHUB_RAW_URL}/$source" -o "$target"
    else
        wget -q "${GITHUB_RAW_URL}/$source" -O "$target"
    fi
}

usage() {
    echo "Usage: $0 [--domain api.example.com | --http]"
    echo "A domain enables automatic HTTPS on 443. Use --http to keep port 8080."
}

# Main installation function
main() {
    local domain="" bind_host="0.0.0.0" access_url="http://localhost:8080"
    case "${1:-}" in
        --domain)
            if [ "$#" -ne 2 ] || ! valid_domain "$2"; then
                print_error 'Enter a DNS name such as api.example.com, without https:// or a port.'
                return 1
            fi
            domain="$2"
            ;;
        --http)
            [ "$#" -eq 1 ] || { usage >&2; return 1; }
            ;;
        -h|--help) usage; return 0 ;;
        '') [ "$#" -eq 0 ] || { usage >&2; return 1; } ;;
        *) usage >&2; return 1 ;;
    esac
    echo ""
    echo "=========================================="
    echo "  Sub2API Deployment Preparation"
    echo "=========================================="
    echo ""

    # Check if openssl is available
    if ! command_exists openssl; then
        print_error "openssl is not installed. Please install openssl first."
        exit 1
    fi

    if ! command_exists curl && ! command_exists wget; then
        print_error "Neither curl nor wget is installed. Please install one of them."
        exit 1
    fi

    # Never regenerate database credentials or replace an existing deployment.
    local existing
    for existing in .env .env.example Caddyfile.https data postgres_data redis_data \
        compose.{yml,yaml} docker-compose.{yml,yaml} {compose,docker-compose}.override.{yml,yaml}; do
        if [ -e "$existing" ] || [ -L "$existing" ]; then
            print_error "Existing deployment file or directory: $existing. Run the installer in a new directory."
            return 1
        fi
    done

    if [ "$#" -eq 0 ] && [ -t 0 ]; then
        read -r -p '域名（自动启用 HTTPS:443，直接回车使用 HTTP:8080）：' domain || return 1
        if [ -n "$domain" ] && ! valid_domain "$domain"; then
            print_error 'Enter a DNS name such as api.example.com, without https:// or a port.'
            return 1
        fi
    fi
    if [ -n "$domain" ]; then
        bind_host="127.0.0.1"
        access_url="https://$domain"
        print_info 'Point the domain to this host and allow inbound TCP 80 and 443 before starting.'
        download_file Caddyfile.https Caddyfile.https
    fi

    print_info "Downloading docker-compose.yml and .env.example..."
    download_file docker-compose.local.yml docker-compose.yml
    download_file .env.example .env.example

    # Generate .env file with auto-generated secrets
    print_info "Generating secure secrets..."
    echo ""

    # Generate secrets
    JWT_SECRET=$(generate_secret)
    TOTP_ENCRYPTION_KEY=$(generate_secret)
    POSTGRES_PASSWORD=$(generate_secret)
    ADMIN_EMAIL=$(generate_admin_email)

    # Create private configuration before writing generated secrets.
    (umask 077; cp .env.example .env)
    chmod 600 .env

    # Update .env with generated secrets (cross-platform compatible)
    if sed --version >/dev/null 2>&1; then
        # GNU sed (Linux)
        sed -i "s/^BIND_HOST=.*/BIND_HOST=${bind_host}/" .env
        sed -i "s/^JWT_SECRET=.*/JWT_SECRET=${JWT_SECRET}/" .env
        sed -i "s/^TOTP_ENCRYPTION_KEY=.*/TOTP_ENCRYPTION_KEY=${TOTP_ENCRYPTION_KEY}/" .env
        sed -i "s/^POSTGRES_PASSWORD=.*/POSTGRES_PASSWORD=${POSTGRES_PASSWORD}/" .env
        sed -i "s/^ADMIN_EMAIL=.*/ADMIN_EMAIL=${ADMIN_EMAIL}/" .env
    else
        # BSD sed (macOS)
        sed -i '' "s/^BIND_HOST=.*/BIND_HOST=${bind_host}/" .env
        sed -i '' "s/^JWT_SECRET=.*/JWT_SECRET=${JWT_SECRET}/" .env
        sed -i '' "s/^TOTP_ENCRYPTION_KEY=.*/TOTP_ENCRYPTION_KEY=${TOTP_ENCRYPTION_KEY}/" .env
        sed -i '' "s/^POSTGRES_PASSWORD=.*/POSTGRES_PASSWORD=${POSTGRES_PASSWORD}/" .env
        sed -i '' "s/^ADMIN_EMAIL=.*/ADMIN_EMAIL=${ADMIN_EMAIL}/" .env
    fi

    if [ -n "$domain" ]; then
        printf '\nSUB2API_DOMAIN=%s\nCOMPOSE_PROFILES=https\n' "$domain" >> .env
    fi

    # Create data directories
    print_info "Creating data directories..."
    mkdir -p data postgres_data redis_data
    print_success "Created data directories"

    echo ""

    # Display completion message
    echo "=========================================="
    echo "  Preparation Complete!"
    echo "=========================================="
    echo ""
    echo "Generated secure credentials:"
    echo "  POSTGRES_PASSWORD:     ${POSTGRES_PASSWORD}"
    echo "  JWT_SECRET:            ${JWT_SECRET}"
    echo "  TOTP_ENCRYPTION_KEY:   ${TOTP_ENCRYPTION_KEY}"
    echo "  ADMIN_EMAIL:           ${ADMIN_EMAIL}  (admin login username)"
    echo ""
    print_warning "These credentials have been saved to .env file."
    print_warning "Please keep them secure and do not share publicly!"
    echo ""
    echo "Directory structure:"
    echo "  docker-compose.yml        - Docker Compose configuration"
    echo "  .env                      - Environment variables (generated secrets)"
    echo "  .env.example              - Example template (for reference)"
    if [ -n "$domain" ]; then
        echo "  Caddyfile.https           - HTTPS reverse proxy configuration"
    fi
    echo "  data/                     - Application data (will be created on first run)"
    echo "  postgres_data/            - PostgreSQL data"
    echo "  redis_data/               - Redis data"
    echo ""
    echo "Next steps:"
    echo "  1. (Optional) Edit .env to customize configuration"
    echo "  2. Start services:"
    echo "     docker compose up -d"
    echo ""
    echo "  3. View logs:"
    echo "     docker compose logs -f sub2api"
    echo ""
    echo "  4. Access Web UI:"
    echo "     $access_url"
    echo ""
    print_info "If admin password is not set in .env, it will be auto-generated."
    print_info "Check logs for the generated admin password on first startup."
    echo ""
}

# Run main function
main "$@"

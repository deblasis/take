#!/bin/bash

# Detect shell
detect_shell() {
    if [ -n "$ZSH_VERSION" ]; then
        echo "zsh"
    elif [ -n "$BASH_VERSION" ]; then
        echo "bash"
    else
        echo "unknown"
    fi
}

# Get shell config file path
get_shell_config() {
    local shell=$1
    case $shell in
        "zsh")
            echo "$HOME/.zshrc"
            ;;
        "bash")
            if [ -f "$HOME/.bashrc" ]; then
                echo "$HOME/.bashrc"
            else
                echo "$HOME/.bash_profile"
            fi
            ;;
        *)
            echo ""
            ;;
    esac
}

# Main installation
main() {
    # Detect shell
    shell=$(detect_shell)
    if [ "$shell" = "unknown" ]; then
        echo "Error: Unsupported shell" >&2
        exit 1
    fi

    # Get config file
    config_file=$(get_shell_config "$shell")
    if [ -z "$config_file" ]; then
        echo "Error: Could not determine shell config file" >&2
        exit 1
    fi

    # Check if take function already exists
    if grep -q "^take()" "$config_file" 2>/dev/null; then
        echo "take function already exists in $config_file"
        exit 0
    fi

    # Add take function to shell config
    cat << 'EOF' >> "$config_file"

# take - Create a new directory and change to it, or download and extract archives
take() {
    if [ -z "$1" ]; then
        echo "Usage: take <directory|git-url|archive-url>" >&2
        return 1
    fi

    # Handle URLs with query parameters
    local target="$1"
    if [[ "$target" == *"?"* ]]; then
        target="${target%%\?*}"
    fi

    # Check if it's a URL
    if [[ "$target" =~ ^(https?|ftp|git|ssh|rsync).*$ ]] || [[ "$target" =~ ^[A-Za-z0-9]+@.*$ ]]; then
        # Create a temporary directory for downloads
        local tmpdir=$(mktemp -d)
        trap 'rm -rf "$tmpdir"' EXIT

        # Handle different URL types
        if [[ "$target" =~ \.(tar\.gz|tgz|tar\.bz2|tar\.xz)$ ]]; then
            # Download and extract tarball
            local archive="$tmpdir/archive.tar"
            if ! curl -L "$target" -o "$archive"; then
                echo "Failed to download archive" >&2
                return 1
            fi
            local dir=$(tar -tf "$archive" | head -n1 | cut -d/ -f1)
            if ! tar xf "$archive"; then
                echo "Failed to extract archive" >&2
                return 1
            fi
            cd "$dir" || return 1
        elif [[ "$target" =~ \.zip$ ]]; then
            # Download and extract ZIP
            local archive="$tmpdir/archive.zip"
            if ! curl -L "$target" -o "$archive"; then
                echo "Failed to download archive" >&2
                return 1
            fi
            local dir=$(unzip -l "$archive" | awk 'NR==4{print $4}' | cut -d/ -f1)
            if ! unzip "$archive"; then
                echo "Failed to extract archive" >&2
                return 1
            fi
            cd "$dir" || return 1
        elif [[ "$target" =~ \.git$ ]] || [[ "$target" =~ ^git@ ]]; then
            # Clone git repository
            local repo=$(basename "$target" .git)
            if ! git clone "$target" "$repo"; then
                echo "Failed to clone repository" >&2
                return 1
            fi
            cd "$repo" || return 1
        else
            echo "Unsupported URL format" >&2
            return 1
        fi
    else
        # Create and change to directory
        mkdir -p "$target" && cd "$target"
    fi
}
EOF

    echo "Installed take function to $config_file"
    echo "Please restart your shell or run: source $config_file"
}

main
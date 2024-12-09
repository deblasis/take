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

# take - Create a new directory and change to it
take() {
    if [ -z "$1" ]; then
        echo "Usage: take <directory or git-url>" >&2
        return 1
    fi
    take_result=$(take-cli "$1")
    if [ $? -eq 0 ]; then
        cd "$take_result"
    else
        echo "$take_result" >&2
        return 1
    fi
}
EOF

    echo "Installed take function to $config_file"
    echo "Please restart your shell or run: source $config_file"
}

main
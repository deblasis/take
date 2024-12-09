#compdef take

_take() {
    local -a opts
    opts=(
        '-depth[Git clone depth (0 for full clone)]:depth:(1 5 10)'
        '-force[Force operation even if directory exists]'
        '-version[Show version information]'
    )

    _arguments -s \
        "${opts[@]}" \
        '*:directory or git URL:->target'

    case $state in
        target)
            if [[ $words[CURRENT] == git@* || $words[CURRENT] == https://* ]]; then
                # URL completion
                local -a git_urls
                git_urls=(${(f)"$(git remote -v 2>/dev/null | awk '{print $2}' | sort -u)"})
                _describe 'git URL' git_urls
            else
                # Directory completion
                _path_files -/
            fi
            ;;
    esac
}

_take "$@"
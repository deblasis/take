# bash completion for take

_take() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="-depth -force -version"

    case "${prev}" in
        -depth)
            COMPREPLY=( $(compgen -W "1 5 10" -- ${cur}) )
            return 0
            ;;
        take)
            # Complete directories and git URLs
            if [[ ${cur} == git@* || ${cur} == https://* ]]; then
                # URL completion
                COMPREPLY=( $(compgen -W "$(git remote -v 2>/dev/null | awk '{print $2}' | sort -u)" -- ${cur}) )
            else
                # Directory completion
                COMPREPLY=( $(compgen -d -- ${cur}) )
            fi
            return 0
            ;;
        *)
            ;;
    esac

    if [[ ${cur} == -* ]] ; then
        COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
        return 0
    fi
}

complete -F _take take
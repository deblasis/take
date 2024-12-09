complete -c take -l depth -d 'Git clone depth (0 for full clone)' -xa '1 5 10'
complete -c take -l force -d 'Force operation even if directory exists'
complete -c take -l version -d 'Show version information'

# Directory completion
complete -c take -a '(__fish_complete_directories)'

# Git URL completion
complete -c take -a '(git remote -v 2>/dev/null | string replace -r "\s+.*" "" | sort -u)' -n 'string match -r "^(git@|https://)" (commandline -ct)'
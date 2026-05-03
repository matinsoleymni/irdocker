# bash completion for irdocker
_irdocker_completions()
{
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="nginx postgres redis help version"

    if [[ ${cur} == -* ]] ; then
        COMPREPLY=( $(compgen -W "--help --version" -- ${cur}) )
        return 0
    fi

    COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
    return 0
}

complete -F _irdocker_completions irdocker

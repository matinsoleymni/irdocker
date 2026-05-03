# bash completion for irdocker(1) — portable, pgsync style

_irdocker()
{
    local cur prev opts cmds
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    cmds="help --help -h list ls add remove rm reset check version --version"

    # If first arg, complete commands or files
    if [[ $COMP_CWORD -eq 1 ]]; then
        COMPREPLY=( $(compgen -W "$cmds" -- "$cur") )
        [[ ${#COMPREPLY[@]} -eq 0 ]] && COMPREPLY=( $(compgen -f -- "$cur") )
        return 0
    fi

    case "${COMP_WORDS[1]}" in
        add)
            # irdocker add <name> <host>
            if [[ $COMP_CWORD -eq 2 ]]; then
                COMPREPLY=()
            elif [[ $COMP_CWORD -eq 3 ]]; then
                COMPREPLY=()
            fi
            ;;
        remove|rm)
            # irdocker remove <host>
            COMPREPLY=()
            ;;
        check)
            # irdocker check <image[:tag]>
            COMPREPLY=()
            ;;
        list|ls|reset|help|version|--version|--help|-h)
            COMPREPLY=()
            ;;
        *)
            # irdocker <image[:tag]> or <compose-file.yaml>
            if [[ "$cur" == -* ]]; then
                COMPREPLY=( $(compgen -W "--help --version" -- "$cur") )
            else
                COMPREPLY=()
            fi
            ;;
    esac
}

complete -F _irdocker irdocker

package shell

import "fmt"

func Init(name string) (string, error) {
	switch name {
	case "bash":
		return `__backhere_last_pwd="$PWD"
__backhere_prompt_command() {
  if [[ "${__backhere_last_pwd-}" != "$PWD" ]]; then
    __backhere_last_pwd="$PWD"
    command backhere check "$PWD"
  fi
}
if [[ "$(declare -p PROMPT_COMMAND 2>/dev/null)" == "declare -a"* ]]; then
  if [[ " ${PROMPT_COMMAND[*]} " != *" __backhere_prompt_command "* ]]; then
      PROMPT_COMMAND+=(__backhere_prompt_command)
  fi
else
  case ";${PROMPT_COMMAND-};" in
    *";__backhere_prompt_command;"*) ;;
    *)
      PROMPT_COMMAND="${PROMPT_COMMAND:+$PROMPT_COMMAND; }__backhere_prompt_command"
      ;;
  esac
fi`, nil
	case "zsh":
		return `typeset -g __backhere_last_pwd="$PWD"
__backhere_chpwd() {
  if [[ "${__backhere_last_pwd-}" != "$PWD" ]]; then
    __backhere_last_pwd="$PWD"
    command backhere check "$PWD"
  fi
}
autoload -Uz add-zsh-hook
add-zsh-hook chpwd __backhere_chpwd`, nil
	case "fish":
		return `set -g __backhere_last_pwd $PWD
functions -e __backhere_on_pwd 2>/dev/null
function __backhere_on_pwd --on-variable PWD
    if test "$__backhere_last_pwd" != "$PWD"
        set -g __backhere_last_pwd $PWD
        command backhere check "$PWD"
    end
end`, nil
	default:
		return "", fmt.Errorf("shell não suportado: %s (use bash, zsh ou fish)", name)
	}
}

# Zsh UX Guide for Flux

This guide styles terminal input and prompt behavior in zsh.

Note: Flux can color only Flux output. The color of text you type is controlled by your shell prompt and line editor settings.

## 1) Make the command prompt cyan

Add this to ~/.zshrc:

```sh
autoload -U colors && colors
PROMPT='%F{cyan}%n@%m%f %F{white}%1~%f %# '
```

Reload shell:

```sh
source ~/.zshrc
```

## 2) Highlight valid commands in cyan while typing

Install syntax highlighting if needed:

```sh
brew install zsh-syntax-highlighting
```

Add this near the end of ~/.zshrc:

```sh
source /opt/homebrew/share/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh
ZSH_HIGHLIGHT_STYLES[command]='fg=cyan,bold'
```

Reload shell:

```sh
source ~/.zshrc
```

## 3) Add lightweight prompt animation

A simple animated spinner can be shown in the right prompt while a command runs.
Add this to ~/.zshrc:

```sh
typeset -gi FLUX_SPINNER_IDX=0
typeset -ga FLUX_SPINNER_FRAMES
FLUX_SPINNER_FRAMES=("-" "\\" "|" "/")

function flux_precmd() {
  RPROMPT="%F{cyan}ready%f"
}

function flux_preexec() {
  FLUX_SPINNER_IDX=0
}

function flux_periodic() {
  FLUX_SPINNER_IDX=$(( (FLUX_SPINNER_IDX + 1) % ${#FLUX_SPINNER_FRAMES[@]} ))
  RPROMPT="%F{cyan}${FLUX_SPINNER_FRAMES[$((FLUX_SPINNER_IDX + 1))]} running%f"
  zle && zle reset-prompt
}

autoload -Uz add-zsh-hook
add-zsh-hook precmd flux_precmd
add-zsh-hook preexec flux_preexec
PERIOD=1
add-zsh-hook periodic flux_periodic
```

Reload shell:

```sh
source ~/.zshrc
```

## 4) Flux command shortcuts

Use shorter Flux commands:

```sh
flux input.md pdf
flux c input.md pdf
flux d
flux lf
flux i input.pdf
```

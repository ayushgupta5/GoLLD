# Command History Recommendations Setup Guide
## दूसरे Mac Laptop पर Command History से Recommendations Enable करने के लिए

### विशेषताएं (Features)
- जो commands आपने पहले use किये हैं, वो automatically suggest होंगे
- Arrow keys से history search होगी
- Fuzzy search से commands ढूंढ सकेंगे
- Real-time autocomplete suggestions

---

## Step 1: Homebrew Install करें (अगर पहले से नहीं है)

```bash
# Check if Homebrew is installed
which brew

# If not installed, install it:
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

---

## Step 2: zsh-autosuggestions Install करें

यह plugin आपके पुराने commands को देखकर automatically suggestions देगा।

```bash
# Install zsh-autosuggestions
brew install zsh-autosuggestions

# Or using Oh My Zsh (if you have it):
git clone https://github.com/zsh-users/zsh-autosuggestions ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-autosuggestions
```

---

## Step 3: zsh-syntax-highlighting Install करें (Optional but recommended)

```bash
brew install zsh-syntax-highlighting
```

---

## Step 4: fzf Install करें (Fuzzy Finder for Command History)

यह सबसे powerful tool है command history search के लिए।

```bash
# Install fzf
brew install fzf

# Install useful key bindings and fuzzy completion
$(brew --prefix)/opt/fzf/install
```

**What fzf does:**
- Press `Ctrl+R` → Search through your entire command history with fuzzy search
- Press `Ctrl+T` → Find files in current directory
- Press `Alt+C` → Change directory with fuzzy search

---

## Step 5: atuin Install करें (Advanced History Management - Optional)

यह एक modern command history tool है जो सभी commands को sync करता है।

```bash
# Install atuin
brew install atuin

# Initialize atuin
atuin init zsh

# Import existing history
atuin import auto
```

---

## Step 6: अपनी .zshrc File Configure करें

अपने दूसरे Mac पर, `~/.zshrc` file में ये lines add करें:

```bash
# Open .zshrc in editor
nano ~/.zshrc
# OR
vim ~/.zshrc
# OR
code ~/.zshrc  # If using VS Code
```

**Add these configurations:**

```bash
# ============================================
# COMMAND HISTORY SETTINGS
# ============================================

# Increase history size
HISTFILE=~/.zsh_history
HISTSIZE=50000
SAVEHIST=50000

# Share history between all sessions
setopt SHARE_HISTORY

# Append to history file instead of overwriting
setopt APPEND_HISTORY

# Save timestamp and duration
setopt EXTENDED_HISTORY

# Don't save duplicate commands
setopt HIST_IGNORE_DUPS
setopt HIST_IGNORE_ALL_DUPS
setopt HIST_FIND_NO_DUPS

# Don't save commands starting with space
setopt HIST_IGNORE_SPACE

# Remove extra spaces from commands before saving
setopt HIST_REDUCE_BLANKS

# ============================================
# ZSH AUTOSUGGESTIONS
# ============================================

# Load zsh-autosuggestions (Homebrew installation)
source $(brew --prefix)/share/zsh-autosuggestions/zsh-autosuggestions.zsh

# Suggestion color (grey)
ZSH_AUTOSUGGEST_HIGHLIGHT_STYLE='fg=240'

# Accept suggestion with right arrow or Ctrl+Space
bindkey '^ ' autosuggest-accept  # Ctrl+Space to accept
bindkey '^[[C' forward-char      # Right arrow to move forward

# ============================================
# ZSH SYNTAX HIGHLIGHTING
# ============================================

# Load syntax highlighting (Homebrew installation)
source $(brew --prefix)/share/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh

# ============================================
# FZF - FUZZY FINDER
# ============================================

# Set up fzf key bindings and fuzzy completion
[ -f ~/.fzf.zsh ] && source ~/.fzf.zsh

# Or if installed via Homebrew:
source <(fzf --zsh)

# Better default options for fzf
export FZF_DEFAULT_OPTS='
  --height 40% 
  --layout=reverse 
  --border 
  --preview "echo {}" 
  --preview-window=down:3:wrap
'

# Ctrl+R for command history search
export FZF_CTRL_R_OPTS="
  --preview 'echo {}' 
  --preview-window down:3:wrap
  --bind 'ctrl-y:execute-silent(echo -n {2..} | pbcopy)+abort'
  --header 'Press CTRL-Y to copy command to clipboard'
"

# ============================================
# ATUIN (Optional - Advanced)
# ============================================

# Uncomment if you installed atuin
# eval "$(atuin init zsh)"

# ============================================
# CUSTOM ALIASES FOR HISTORY
# ============================================

# Quick history search
alias h='history'
alias hg='history | grep'

# Show most used commands
alias histop='history | awk "{print \$2}" | sort | uniq -c | sort -rn | head -20'

# Clear history (use carefully!)
alias histclear='echo "" > ~/.zsh_history && history -c && exec zsh'
```

---

## Step 7: Changes Apply करें

```bash
# Reload your .zshrc
source ~/.zshrc

# OR restart your terminal
```

---

## Usage Guide - कैसे Use करें

### 1. **Auto-suggestions (zsh-autosuggestions)**
- जैसे ही आप typing शुरू करेंगे, grey color में suggestion दिखेगा
- **Right Arrow** या **End** key press करें suggestion accept करने के लिए
- **Ctrl+Space** से भी accept कर सकते हैं

### 2. **Command History Search (fzf)**
- **Ctrl+R** press करें
- अपनी command का कोई भी part type करें (fuzzy search)
- Arrow keys से navigate करें
- Enter press करें command select करने के लिए

### 3. **Up/Down Arrow Keys**
- Up arrow: पिछली command
- Down arrow: अगली command
- अगर कुछ type किया है, तो matching commands ही show होंगी

### 4. **History Command**
```bash
# सभी history देखें
history

# Last 20 commands
history -20

# Specific command search
history | grep "docker"

# Most used commands
histop
```

---

## Step 8: History को Sync/Transfer करें (Optional)

अगर आप इस Mac की history को दूसरे Mac पर transfer करना चाहते हैं:

```bash
# इस Mac से (जहां history है):
scp ~/.zsh_history username@other-mac-ip:~/zsh_history_backup

# दूसरे Mac पर:
cat ~/zsh_history_backup >> ~/.zsh_history
```

---

## Troubleshooting

### अगर autosuggestions काम नहीं कर रहे:
```bash
# Check if plugin is loaded
echo $ZSH_AUTOSUGGEST_HIGHLIGHT_STYLE

# Reinstall
brew reinstall zsh-autosuggestions
source ~/.zshrc
```

### अगर fzf काम नहीं कर रहा:
```bash
# Reinstall key bindings
$(brew --prefix)/opt/fzf/install --all

# Reload
source ~/.zshrc
```

### अगर history save नहीं हो रही:
```bash
# Check permissions
ls -la ~/.zsh_history

# Fix permissions
chmod 600 ~/.zsh_history
```

---

## Advanced: Oh My Zsh के साथ (Optional)

अगर आप Oh My Zsh use करते हैं:

```bash
# Install Oh My Zsh
sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"

# In ~/.zshrc, add these plugins:
plugins=(
  git
  zsh-autosuggestions
  zsh-syntax-highlighting
  history
  history-substring-search
  fzf
)
```

---

## Quick Reference Card

| Shortcut | Action |
|----------|--------|
| `Ctrl+R` | Fuzzy search history (fzf) |
| `Right Arrow` | Accept suggestion |
| `Ctrl+Space` | Accept suggestion (alternate) |
| `Up/Down` | Browse history |
| `Ctrl+P/N` | Previous/Next (alternate) |
| `history` | Show all history |
| `!!` | Repeat last command |
| `!$` | Last argument of previous command |
| `!*` | All arguments of previous command |
| `!abc` | Last command starting with 'abc' |

---

## Summary - सारांश

1. ✅ Homebrew install करें
2. ✅ zsh-autosuggestions install करें → Real-time suggestions
3. ✅ fzf install करें → Powerful history search (Ctrl+R)
4. ✅ .zshrc configure करें → History settings optimize
5. ✅ Terminal restart करें
6. ✅ Commands type करना शुरू करें और suggestions देखें!

अब आपके दूसरे Mac पर commands automatically recommend होने लगेंगी! 🎉


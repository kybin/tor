# tor 

Tor is a simple terminal text editor. 

It's not stable or well coded, but usable for me. 

### Key Binding

Tor has different key binding set from other editors. 

Basically you could think `Ctrl` is for action, `Alt` is for move, `Shift` is for selection.

`Esc` or `Ctrl+K` terminates any special mode and return to normal mode.

### How to use

#### Basic
- New file : `$ tor -new filename.ext`
- Save : `Ctrl+S`
- Undo : `Ctrl+Z`
- Quit : `Ctrl+W`

#### Move
- Left : `Alt+J`
- Right : `Alt+L`
- Up : `Alt+I`
- Down : `Alt+K`
- Prev Word Edge : `Alt+M`
- Next Word Edge : `Alt+.`
- Goto Line : `Ctrl+G`
- New Line : `Ctrl+N`
- Indent Line : `Ctrl+O`
- Unindent Line : `Ctrl+U`
- Page Up : `Alt+W`
- Page Down : `Alt+S`
- Home : `Alt+Q`
- End : `Alt+A`
- Bigining Of Contents : `Alt+U` 
- End Of Contents : `Alt+O` 
- Prev Head Line : `Alt+9`
- Next Head Line : `Alt+0`

#### Select
- Select Line : `Ctrl+L`
- Select Mode : Shift+MoveAction
  - ex) Select Word :`Shift+Alt+.`

#### Copy, Paste
- Copy : `Ctrl+C`
- Paste : `Ctrl+V`

#### Find, Replace
- Find Mode : `Ctrl+F` 
- Find Next : `Ctrl+D`
- Find Prev : `Ctrl+B`
- Replace Mode : `Ctrl+R`
- Replace : `Ctrl+J`
- Cancel Input Mode : `Ctrl+K`

#### Other
- ...And several other key maps, but they may changed frequently.

#### Install

Install tor as other go programs.

`go get -u github.com/kybin/tor`


It is not must but recommended to install goimports when you programming in go.

- goimports : `go get -u golang.org/x/tools/cmd/goimports`

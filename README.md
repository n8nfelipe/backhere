# BackHere

> Deixe um lembrete onde você está. Receba-o quando voltar.

```console
$ cd ~/projects/my-api

$ backhere add "Check payment timeout"
✓ Lembrete salvo aqui.

# alguns dias depois...

$ cd ~/projects/my-api

📌 BackHere
1. Check payment timeout
```

BackHere é uma CLI pequena e local. O diretório é o contexto: não há contas,
servidor, datas ou notificações agendadas.

## Instalação

Requer Go 1.25 ou mais recente:

```sh
go install github.com/n8nfelipe/backhere@latest
```

Ou compile este repositório com `make build` e copie `bin/backhere` para um
diretório presente no `PATH`.

## Uso

```sh
backhere add Corrigir autenticação
backhere list
backhere done 1
backhere clear
```

`done 1` usa o número mostrado por `list`. `clear` pede confirmação e afeta
somente os lembretes ativos do diretório atual.

## Integração com o shell

Adicione uma linha ao arquivo de configuração do seu shell:

```sh
# ~/.bashrc
eval "$(backhere init bash)"

# ~/.zshrc
eval "$(backhere init zsh)"
```

Para Fish, adicione a `~/.config/fish/config.fish`:

```fish
backhere init fish | source
```

O script chama `backhere check` somente quando `$PWD` realmente muda. O último
diretório fica em uma variável da sessão: não há polling, daemon ou escrita
extra no banco. Sair de B, visitar C e voltar a B mostra os lembretes novamente.

## Dados e caminhos

O SQLite fica no diretório de dados local do sistema:

- Linux: `$XDG_DATA_HOME/backhere/backhere.db`, ou `~/.local/share/backhere/backhere.db`;
- macOS: `~/Library/Application Support/BackHere/backhere.db`;
- Windows: `%LOCALAPPDATA%\BackHere\backhere.db`.

BackHere transforma caminhos em absolutos e resolve symlinks existentes. Assim,
um link e seu destino compartilham o mesmo contexto. No Windows, o caminho
normalizado também é comparado sem diferença entre maiúsculas e minúsculas.

## Desenvolvimento

```sh
make test
make build
```

O driver `modernc.org/sqlite` é puro Go: SQLite já faz parte do binário e CGO
não é necessário.

## Desinstalação

Remova o binário, apague a linha de integração do shell e, se quiser apagar os
dados, remova a pasta `BackHere` indicada acima.

Licença [MIT](LICENSE).

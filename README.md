# md-notes (`mdn`)

> Editor e visualizador modal de Markdown em terminal (TUI) moderno, de alta performance e multiplataforma, construído em Go com [Bubble Tea](https://github.com/charmbracelet/bubbletea) e [Lipgloss](https://github.com/charmbracelet/lipgloss).

---

## Recursos Principais

- **Multiplataforma Nativo**: Executáveis estáticos e independentes para **macOS**, **Linux** e **Windows**.
- **Inicialização Instantânea**: Menos de 50 milissegundos de boot e consumo mínimo de memória (< 25 MB).
- **Edição Modal Estilo Vim**: Modos Normal, Insert, Visual (linhas/blocos) e Command (`:w`, `:q`, `:s/de/para/g`).
- **Renderização e Realce Markdown em Tempo Real**: Sintaxe colorida, tabelas dinâmicas autoajustáveis e blocos de código.
- **Configuração TOML e Temas Dinâmicos**: Dracula, Nord, Catppuccin, Monokai ou paleta customizada com hot-reload automático.
- **Integração Unix / CLI**: Suporte a pipe (`cat file.md | mdn`), scratchpad temporário e edição de arquivos existentes.

---

## Seleção, cópia e colagem

Arraste o botão esquerdo do mouse sobre o texto para selecionar. A seleção continua visível após soltar; `Esc` a cancela. `Ctrl+C` copia o Markdown original da seleção, incluindo suas quebras de linha. Sem seleção, `Ctrl+C` tenta sair e protege alterações não salvas com `E37`.

Um clique com o botão direito dentro do documento copia a seleção ativa. Sem seleção, cola o texto do clipboard na posição clicada; na área vazia abaixo da última linha, cola ao fim do documento. A colagem pode ser desfeita com um único `u`. Cliques na barra de status ou no painel de configurações não alteram o documento.

No modo INSERT, `Ctrl+V` cola texto simples do clipboard. Com uma seleção ativa, `Ctrl+V` substitui o trecho selecionado; no modo NORMAL sem seleção, mantém o comando Vim de bloco visual. Colagem delimitada pelo terminal também insere o texto literalmente em INSERT ou sobre uma seleção, como uma única etapa de desfazer. Cada colagem aceita até 10.000 caracteres Unicode e 200 linhas.

Os atalhos dependem do terminal: o `mdn` só recebe `Cmd+C` quando o terminal o encaminha como `Ctrl+C`, e só recebe `Cmd+V` quando o terminal envia uma colagem delimitada ou o mapeia para `Ctrl+V`. Se o terminal ou o `tmux` capturar mouse ou atalhos, configure seu encaminhamento; no `tmux`, habilite o mouse com `set -g mouse on`.

---

## Compilação e Build Multiplataforma

O projeto conta com infraestrutura completa de build para as 3 plataformas (macOS, Linux e Windows) tanto em arquitetura x86_64 (`amd64`) quanto ARM64 (`arm64`).

### Compilação Rápida (Plataforma Atual)

```bash
make build
./bin/mdn --version
```

### Compilar para Todas as Plataformas (macOS, Linux, Windows)

```bash
# Via Makefile
make build-all

# Ou usando o script em Go (funciona em qualquer OS sem dependências extras):
go run scripts/build.go
```

Os binários compilados serão gerados em `bin/`:
- `bin/darwin_arm64/mdn` (macOS Apple Silicon)
- `bin/darwin_amd64/mdn` (macOS Intel)
- `bin/linux_amd64/mdn` (Linux x86_64)
- `bin/linux_arm64/mdn` (Linux ARM64)
- `bin/windows_amd64/mdn.exe` (Windows x64)
- `bin/windows_arm64/mdn.exe` (Windows ARM64)

### Gerar Pacotes de Release (.tar.gz, .zip e Checksums SHA256)

```bash
# Gera os arquivos comprimidos em dist/ com checksums.txt
make package

# Ou com o builder Go:
go run scripts/build.go -package
```

Para detalhes completos sobre opções de compilação, `lipo` para Universal Binary no macOS, GoReleaser e fluxos de CI/CD, consulte a documentação detalhada em [docs/BUILD.md](docs/BUILD.md).

---

## Testes

```bash
# Rodar todos os testes unitários
make test

# Rodar testes com relatório de cobertura HTML
make test-coverage
```

---

## Licença

MIT

# Especificação de Casos de Teste: F02. Configuration & Theme Engine

## Sumário de Cobertura

| Tipo de Teste | Quantidade | Foco / Alvo |
|---|---|---|
| Unitário | 6 | Resolução de caminhos, integridade de temas, overrides de cores e compilação Lipgloss |
| Integração | 6 | Parser TOML, fallback de sintaxe, geração de defaults, hot-reload `fsnotify` e contratos de estilo |
| E2E / CLI | 3 | Inicialização do executável com arquivos de configuração em disco e alternância de temas |

---

## Matriz de Rastreabilidade

| Critério de Aceitação (PRD Seção 9) | Casos de Teste |
|---|---|
| O editor carrega o arquivo `~/.config/md-notes/config.toml` caso exista no sistema | TC01, TC07 |
| Na ausência do arquivo de configuração, o editor cria o `config.toml` padrão documentado e aplica o tema `default-dark` | TC02, TC08 |
| Configurar um dos temas pré-definidos (`dracula`, `nord`, `catppuccin-mocha`, `catppuccin-macchiato`, `monokai`, `default-dark`, `default-light`) aplica a paleta correspondente | TC03, TC09 |
| Cores customizadas em Hex (`#RRGGBB`) definidas no TOML sobrepõem corretamente as cores dos tokens de Markdown e interface | TC04, TC10 |
| Erros de sintaxe no arquivo TOML são tratados com fallback seguro para o tema padrão sem travar o aplicativo | TC05, TC11 |
| Cross-Feature: Estilos de Theme Engine fornecem tokens uniformes para F04, F05, F06 e F07 | TC06, TC12 |
| Full Scope: Hot-reload em disco via `fsnotify` e detecção de modo claro/escuro (`theme = "auto"`) | TC13, TC14, TC15 |

---

## Casos de Teste Unitários

### TC01 — Resolução de Diretórios Padrão XDG / AppData
- **Alvo:** `internal/config/config.go` (`GetConfigFilePath`)
- **Tipo:** unit
- **Pré-condições:** Nenhuma.
- **Passos:**
  1. Invocar `GetConfigFilePath()`.
  2. Avaliar o caminho absoluto retornado em função do sistema operacional atual.
- **Resultado Esperado:** Retorna terminação `.config/md-notes/config.toml` em Linux/macOS ou `AppData\Roaming\md-notes\config.toml` em Windows.
- **Cobre:** Critério PRD: carregamento em diretórios padrão XDG.

### TC02 — Validação de Estrutura e Valores Padrão da Configuração
- **Alvo:** `internal/config/config.go` (`DefaultConfig`)
- **Tipo:** unit
- **Pré-condições:** Nenhuma.
- **Passos:**
  1. Invocar `DefaultConfig()`.
  2. Checar campos `Theme`, `Editor.TabSize` e `Editor.LineNumbers`.
- **Resultado Esperado:** `Theme == "default-dark"`, `Editor.TabSize == 4` e `Editor.LineNumbers == true`.
- **Cobre:** Definição dos valores padrão recomendados da aplicação.

### TC03 — Integridade das 7 Paletas de Cores Embutidas
- **Alvo:** `internal/theme/palette.go` (`GetPalette`)
- **Tipo:** unit
- **Pré-condições:** Nenhuma.
- **Passos:**
  1. Iterar sobre os 7 temas: `dracula`, `nord`, `catppuccin-mocha`, `catppuccin-macchiato`, `monokai`, `default-dark`, `default-light`.
  2. Validar que cada paleta retorna `found == true`.
  3. Verificar que nenhum token de cor (`H1` a `H6`, `Muted`, `TableBorder`, etc.) está vazio.
- **Resultado Esperado:** Todos os 7 temas estão presentes e com paletas completas e válidas.
- **Cobre:** Critério PRD: suporte aos 7 temas pré-definidos.

### TC04 — Compilação e Sobreposição de Cores Granulares (Hex / ANSI)
- **Alvo:** `internal/theme/compiler.go` (`CompileTheme`)
- **Tipo:** unit
- **Pré-condições:** Paleta base `dracula` e struct `ColorOverrides` com `H1 = "#FF0000"` e `TableBorder = "#00FF00"`.
- **Passos:**
  1. Invocar `CompileTheme(palette, overrides)`.
  2. Inspecionar as instâncias `lipgloss.Style` de `H1` e `TableBorder`.
- **Resultado Esperado:** O estilo `H1` reflete a cor vermelha customizada sobrepondo o valor padrão do Dracula, enquanto os demais tokens permanecem com os valores originais do Dracula.
- **Cobre:** Critério PRD: customização via Hex/ANSI sobrepondo cores da paleta.

### TC05 — Detecção de Modo Claro/Escuro do Terminal (`COLORFGBG`)
- **Alvo:** `internal/theme/detect.go` (`ResolveThemeName`)
- **Tipo:** unit
- **Pré-condições:** Variável de ambiente `COLORFGBG` configurada para `"0;15"` (fundo claro) e `"15;0"` (fundo escuro).
- **Passos:**
  1. Chamar `ResolveThemeName("auto")` com `COLORFGBG="0;15"`.
  2. Chamar `ResolveThemeName("auto")` com `COLORFGBG="15;0"`.
- **Resultado Esperado:** Retorna `"default-light"` para fundo claro e `"default-dark"` para fundo escuro.
- **Cobre:** Full Scope PRD: suporte a Dark/Light mode detection do terminal.

### TC06 — Performance de Compilação de Estilos (< 5 ms)
- **Alvo:** `internal/theme/compiler.go` (`CompileTheme`)
- **Tipo:** unit (benchmark / latência)
- **Pré-condições:** Nenhuma.
- **Passos:**
  1. Medir o tempo de execução de `CompileTheme` em 1.000 iterações.
- **Resultado Esperado:** Tempo de compilação individual inferior a 5 milissegundos por chamada (média < 1 ms).
- **Cobre:** Capacidade PRD: compilação de estilos em menos de 5 ms.

---

## Casos de Teste de Integração

### TC07 — Carregamento de Arquivo `config.toml` Existente
- **Alvo:** `internal/config/loader.go` (`LoadConfig`)
- **Tipo:** integration
- **Pré-condições:** Arquivo `config.toml` criado em diretório temporário com `theme = "nord"` e `tab_size = 2`.
- **Passos:**
  1. Invocar o carregador de configuração apontando para o arquivo de teste.
  2. Inspecionar o objeto `Config` retornado.
- **Resultado Esperado:** `Config.Theme == "nord"` e `Config.Editor.TabSize == 2`.
- **Cobre:** Critério PRD: editor carrega arquivo `config.toml` caso exista.

### TC08 — Criação Automática do Arquivo Padrão Documentado
- **Alvo:** `internal/config/loader.go` (`EnsureDefaultConfigFile`)
- **Tipo:** integration
- **Pré-condições:** Diretório de configuração vazio.
- **Passos:**
  1. Chamar `EnsureDefaultConfigFile(caminhoArquivo)`.
  2. Verificar criação física do arquivo no disco e ler seu conteúdo em texto.
- **Resultado Esperado:** O arquivo é criado contendo comentários explicativos sobre os 7 temas, seções `[editor]` e `[colors]`, e parâmetros padrão.
- **Cobre:** Critério PRD: na ausência do arquivo, cria `config.toml` padrão documentado.

### TC09 — Aplicação de Temas Pré-definidos nos Tokens de Renderização
- **Alvo:** `internal/config/loader.go` & `internal/theme/compiler.go`
- **Tipo:** integration
- **Pré-condições:** Configurações com cada um dos temas (`monokai`, `catppuccin-mocha`, etc.).
- **Passos:**
  1. Carregar configuração e compilar tema.
  2. Renderizar string de teste com `theme.H1.Render("# Título")`.
- **Resultado Esperado:** A string renderizada contém as sequências de escape ANSI corretas correspondentes à paleta do tema ativo.
- **Cobre:** Critério PRD: configuração de tema pré-definido aplica paleta correspondente.

### TC10 — Preservação e Precedência de Cores Customizadas
- **Alvo:** `internal/config/loader.go` & `internal/theme/compiler.go`
- **Tipo:** integration
- **Pré-condições:** Arquivo TOML definindo `theme = "dracula"` e `colors.code_bg = "#112233"`.
- **Passos:**
  1. Fazer o parse do TOML e compilar `CompiledTheme`.
  2. Inspecionar `CompiledTheme.CodeBlock`.
- **Resultado Esperado:** A cor de fundo do bloco de código é exatamente `#112233`, mantendo as demais cores conforme a paleta Dracula.
- **Cobre:** Critério PRD: cores customizadas em Hex sobrepõem os tokens padrão.

### TC11 — Resiliência a Erros de Sintaxe TOML com Fallback Seguro
- **Alvo:** `internal/config/loader.go` (`LoadConfig`)
- **Tipo:** integration
- **Pré-condições:** Arquivo `config.toml` com sintaxe corrompida (`theme = [invalido_sem_fechar`).
- **Passos:**
  1. Executar `LoadConfig()`.
  2. Verificar o objeto `Config` retornado e os warnings emitidos.
- **Resultado Esperado:** A função não entra em panic nem retorna erro fatal; entrega `Config` com `default-dark` e inclui aviso descritivo do erro de sintaxe.
- **Cobre:** Critério PRD: erros de sintaxe são tratados com fallback seguro sem travar o aplicativo.

### TC12 — Contrato de Exportação de `CompiledTheme` para Outros Módulos
- **Alvo:** `internal/theme/compiler.go` (`CompiledTheme`)
- **Tipo:** integration
- **Pré-condições:** `CompiledTheme` instanciado.
- **Passos:**
  1. Verificar disponibilização pública de todos os campos exigidos por F04 (H1..H6, Muted, Bold, Italic, CodeBlock), F05 (TableBorder, TableHeader, TableCell), F06 (SearchMatch) e F07 (StatusBar, StatusBarMode, LineNumber).
- **Resultado Esperado:** Todos os campos estão acessíveis como `lipgloss.Style` prontos para renderização imediata.
- **Cobre:** Critério Cross-Feature PRD: estilos consumidos e aplicados uniformemente.

---

## Casos de Teste E2E / Hot-Reload

### TC13 — Hot-Reload Dinâmico de Tema via `fsnotify` (Full Scope)
- **Alvo:** `internal/config/watcher.go`
- **Tipo:** e2e / integration
- **Pré-condições:** Programa em execução com watcher ativo no `config.toml` configurado como `dracula`.
- **Passos:**
  1. Sobrescrever o arquivo `config.toml` alterando `theme = "nord"`.
  2. Aguardar evento `ThemeReloadedMsg` despachado após janela de debounce.
  3. Inspecionar o novo `CompiledTheme` ativo.
- **Resultado Esperado:** O tema ativo muda para `nord` em tempo real sem interrupção da sessão.
- **Cobre:** Full Scope PRD: hot-reload do arquivo de configuração ao ser alterado em disco.

### TC14 — Inicialização E2E com Geração de Configuração
- **Alvo:** Executável compilado `mdn`
- **Tipo:** e2e
- **Pré-condições:** Ambiente isolado com `HOME=/tmp/sandbox_e2e_f02`.
- **Passos:**
  1. Executar `./mdn` e sair imediatamente com `:q`.
  2. Verificar existência de `/tmp/sandbox_e2e_f02/.config/md-notes/config.toml`.
- **Resultado Esperado:** O arquivo de configuração padrão é criado automaticamente com sucesso.
- **Cobre:** Critério PRD: criação automática de `config.toml` na primeira execução.

### TC15 — Inicialização E2E com TOML Corrompido
- **Alvo:** Executável compilado `mdn`
- **Tipo:** e2e
- **Pré-condições:** Arquivo `/tmp/sandbox_e2e_f02/.config/md-notes/config.toml` contendo caracteres inválidos.
- **Passos:**
  1. Executar `./mdn` e sair com `:q`.
- **Resultado Esperado:** O editor abre normalmente no tema padrão e encerra com código 0 sem crash.
- **Cobre:** Critério PRD: resiliência a arquivos inválidos em execução real.

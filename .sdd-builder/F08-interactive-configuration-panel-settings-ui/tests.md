# Especificação de Casos de Teste: F08. Interactive Configuration Panel & Settings UI

## Sumário de Cobertura

| Tipo | Quantidade | Alvo / Escopo |
|---|---|---|
| Unitário | 11 | Estado do painel, navegação, validações hex, toggles de editor, persistência atômica e overlay blending |
| Integração | 7 | Comandos Ex, ciclo de live preview, restauração de snapshot, coordenação com watcher e atalho global F2 |

---

## Matriz de Rastreabilidade

| Critério de Aceitação (PRD Seção 9) | Casos de Teste |
|---|---|
| O comando `:c` ou `:config` executado no modo Normal abre o popup de configurações centralizado na tela | TC01 |
| O popup renderiza opções organizadas para alternância de tema e parâmetros do editor (`tab_size`, `line_numbers`, etc.) | TC06, TC07, TC08, TC09 |
| As teclas de navegação (`j`/`k` ou setas cima/baixo) movem a seleção entre os itens configuráveis do painel | TC04, TC05 |
| A navegação pelos temas exibe a lista dos 7 temas incorporados e reflete o tema ativo no editor | TC06 |
| Pressionar `Enter` na confirmação grava as configurações alteradas no arquivo `config.toml` de forma atômica e fecha o modal | TC14 |
| Pressionar `Esc` ou `q` fecha o painel de configurações imediatamente sem persistir alterações | TC13 |
| Em caso de erro de gravação de arquivo no disco, o painel exibe aviso de falha na interface sem travar a sessão | TC16 |
| Em terminais com dimensões menores que 50x14, o editor recusa abrir o modal e emite alerta informativo na linha de status | TC02 |
| Cross-Feature: O comando `:c`/`:config` em F03 inicializa F08 e direciona temporariamente a captura de teclas ao modal | TC01, TC17 |
| Cross-Feature: As opções e paletas de F02 são consumidas e manipuladas pela interface de F08 | TC06, TC10 |
| Cross-Feature: A gravação em F08 persiste o arquivo TOML e notifica F02 e F07 para recarregamento imediato | TC14, TC15 |
| Cross-Feature: O TUI Viewport (F07) centraliza e renderiza a camada visual de sobreposição mantendo o documento íntegro | TC18 |
| Full Scope: Live preview dinâmico no viewport de fundo conforme navega na lista de temas | TC12 |
| Full Scope: Edição interativa de substituições de cores Hex (`#RRGGBB`) com validação | TC10, TC11 |
| Full Scope: Atalho de teclado rápido global `F2` para alternar abertura/fechamento do painel | TC03 |

---

## Casos de Teste Unitários

### TC01 — Execução dos Comandos Ex `:c` e `:config`
- **Alvo:** `internal/vim/command.go` (`ExecuteExCommand`)
- **Tipo:** unit
- **Pré-condições:** Motor Vim inicializado em modo Normal com callback de configuração configurado.
- **Passos:**
  1. Executar `ExecuteExCommand("c", ...)` e inspecionar o resultado retornado.
  2. Executar `ExecuteExCommand("config", ...)` e inspecionar o resultado retornado.
- **Resultado Esperado:** Ambos os comandos retornam `CloseMode: true`, não geram erro `E492` e acionam a execução do callback de abertura do painel.
- **Covers:** Critério PRD: comando `:c`/`:config` abre o popup de configurações.

### TC02 — Rejeição de Abertura em Janela Insuficiente (< 50x14)
- **Alvo:** `internal/app/model.go` (`OpenSettingsMsg`)
- **Tipo:** unit
- **Pré-condições:** Aplicação com terminal redimensionado para 48 colunas por 12 linhas.
- **Passos:**
  1. Enviar mensagem de solicitação de abertura de configurações.
- **Resultado Esperado:** O estado `SettingsOpen` permanece `false`, nenhum diálogo é aberto e a barra de status exibe: `"Dimensão insuficiente (mínimo 50x14) para abrir o painel de configurações"`.
- **Covers:** Critério PRD: recusa abertura em terminais menores que 50x14 com aviso informativo.

### TC03 — Alternância de Abertura via Atalho Global F2
- **Alvo:** `internal/app/model.go` (`Update`)
- **Tipo:** unit
- **Pré-condições:** Aplicação em execução com dimensões 80x24 e painel de configurações fechado.
- **Passos:**
  1. Enviar evento de tecla `F2`.
  2. Verificar se `SettingsOpen` tornou-se `true`.
  3. Enviar novamente o evento de tecla `F2`.
- **Resultado Esperado:** A primeira pressão abre o painel; a segunda pressão fecha o diálogo e cancela qualquer alteração temporária pendente.
- **Covers:** Full Scope: atalho de teclado rápido global `F2`.

### TC04 — Navegação Cíclica Entre Abas
- **Alvo:** `internal/ui/settings.go` (`HandleKey`)
- **Tipo:** unit
- **Pré-condições:** `SettingsState` inicializado com aba ativa `TabThemes`.
- **Passos:**
  1. Pressionar `Tab` ou seta para a direita `Right`.
  2. Verificar a aba selecionada.
  3. Pressionar novamente `Tab`.
  4. Pressionar `Tab` pela terceira vez.
- **Resultado Esperado:** A aba transita ordenadamente: `TabThemes` -> `TabEditor` -> `TabColors` -> `TabThemes`.
- **Covers:** Critério PRD: navegação suave entre seções por teclado.

### TC05 — Navegação Vertical Entre Itens com Limites de Borda
- **Alvo:** `internal/ui/settings.go` (`HandleKey`)
- **Tipo:** unit
- **Pré-condições:** Aba `TabThemes` com 7 temas disponíveis, cursor no índice 0 (`default-dark`).
- **Passos:**
  1. Pressionar `k` ou seta para cima `Up`.
  2. Pressionar `j` ou seta para baixo `Down` 6 vezes até o último tema (`monokai`).
  3. Pressionar `j` mais uma vez.
- **Resultado Esperado:** O cursor não ultrapassa o índice 0 ao mover para cima e permanece no índice 6 (último tema) ao tentar avançar além do limite.
- **Covers:** Critério PRD: teclas `j`/`k` ou setas movem a seleção entre os itens.

### TC06 — Listagem e Seleção dos 7 Temas Incorporados
- **Alvo:** `internal/ui/settings.go`
- **Tipo:** unit
- **Pré-condições:** `SettingsState` inicializado consumindo `theme.BuiltinThemes()`.
- **Passos:**
  1. Consultar a lista de opções de temas carregada no estado.
  2. Mover o cursor para o tema `"dracula"` e pressionar `Enter` ou `Space`.
- **Resultado Esperado:** Contém exatamente os 7 temas (`default-dark`, `default-light`, `dracula`, `nord`, `catppuccin-mocha`, `catppuccin-macchiato`, `monokai`), e a seleção atualiza a preferência de tema ativo para `"dracula"`.
- **Covers:** Critério PRD: exibe a lista dos 7 temas incorporados e reflete o tema ativo.

### TC07 — Alternância de Preferências Booleanas do Editor
- **Alvo:** `internal/ui/settings.go` (`HandleKey`)
- **Tipo:** unit
- **Pré-condições:** Aba `TabEditor` com cursor sobre a opção `line_numbers` (inicialmente `true`).
- **Passos:**
  1. Pressionar `Space` ou `Enter`.
  2. Verificar o novo valor de `line_numbers`.
  3. Pressionar novamente `Space`.
- **Resultado Esperado:** O valor inverte de `true` para `false` no primeiro acionamento e retorna para `true` no segundo acionamento.
- **Covers:** Critério PRD: alternância de parâmetros do editor (`line_numbers`, `word_wrap`, etc.).

### TC08 — Ciclo Sequencial de Tamanho de Tabulação
- **Alvo:** `internal/ui/settings.go` (`HandleKey`)
- **Tipo:** unit
- **Pré-condições:** Cursor sobre a opção `tab_size` (inicialmente 4).
- **Passos:**
  1. Pressionar `Space` ou `Enter`.
  2. Pressionar novamente `Space`.
  3. Pressionar novamente `Space`.
- **Resultado Esperado:** A largura da tabulação transita ciclicamente: `4 -> 8 -> 2 -> 4`.
- **Covers:** Critério PRD: ajuste de `tab_size` (2, 4, 8).

### TC09 — Ajuste com Limites Numéricos de Scrolloff
- **Alvo:** `internal/ui/settings.go` (`HandleKey`)
- **Tipo:** unit
- **Pré-condições:** Cursor sobre o campo `scrolloff` (inicialmente 4).
- **Passos:**
  1. Incrementar até ultrapassar 10.
  2. Decrementar até tentar reduzir abaixo de 0.
- **Resultado Esperado:** O valor incrementa até o teto estrito de 10 e decrementa até o piso estrito de 0 sem valores negativos.
- **Covers:** Critério PRD: ajuste de margem de rolagem `scrolloff`.

### TC10 — Validação e Aplicação de Códigos Hexadecimais
- **Alvo:** `internal/ui/settings.go`
- **Tipo:** unit
- **Pré-condições:** Aba `TabColors` com foco no token `h1`.
- **Passos:**
  1. Pressionar `e` para ativar modo de edição inline.
  2. Digitar `#FF79C6` e pressionar `Enter`.
- **Resultado Esperado:** O valor é validado pela expressão regular `^#[0-9a-fA-F]{6}$`, o buffer inline é fechado e a substituição `Colors.H1` é atualizada para `"#FF79C6"`.
- **Covers:** Full Scope: edição interativa de substituições de cores Hex.

### TC11 — Rejeição de Entradas Hexadecimais Inválidas
- **Alvo:** `internal/ui/settings.go`
- **Tipo:** unit
- **Pré-condições:** Modo de edição inline ativado na aba `TabColors`.
- **Passos:**
  1. Digitar código inválido como `#123`, `FF00FF` ou `#ZZZZZZ`.
  2. Pressionar `Enter`.
- **Resultado Esperado:** A entrada é rejeitada, a cor anterior é preservada e uma mensagem de alerta é exibida: `"Código de cor inválido: use formato #RRGGBB"`.
- **Covers:** Validação sintática robusta de cores customizadas.

---

## Casos de Teste de Integração

### TC12 — Live Preview Dinâmico de Tema no Documento de Fundo
- **Alvo:** `internal/app/model.go`
- **Tipo:** integration
- **Pré-condições:** Editor exibindo documento Markdown com títulos e tema `"default-dark"`. Painel de configurações aberto.
- **Passos:**
  1. Navegar na lista de temas e mudar a seleção para `"dracula"`.
  2. Renderizar um frame do modelo (`m.View()`).
- **Resultado Esperado:** O tema compilado ativo em `m.Theme` passa imediatamente para Dracula e o texto do documento visível nas margens ao redor do diálogo reflete as cores do novo tema em tempo real.
- **Covers:** Full Scope: Live preview dinâmico do tema no viewport de fundo.

### TC13 — Cancelamento Imediato e Restauração Fiel com Esc ou 'q'
- **Alvo:** `internal/app/model.go`
- **Tipo:** integration
- **Pré-condições:** Painel de configurações aberto; tema alterado temporariamente de `"nord"` para `"monokai"` e `line_numbers` alterado para `false`.
- **Passos:**
  1. Pressionar `Esc` ou `q`.
  2. Verificar o estado de `m.Config`, `m.Theme` e fechamento do painel.
- **Resultado Esperado:** O painel se fecha (`SettingsOpen == false`), as preferências temporárias são descartadas e o tema `"nord"` com `line_numbers = true` é integralmente restaurado.
- **Covers:** Critério PRD: pressionar `Esc` ou `q` fecha o painel imediatamente sem persistir alterações.

### TC14 — Persistência Atômica no Arquivo config.toml com Enter
- **Alvo:** `internal/config/saver.go` (`SaveConfigFile`) e `internal/app/model.go`
- **Tipo:** integration
- **Pré-condições:** Diretório de configuração isolado em `t.TempDir()`. Painel aberto com tema alterado para `"nord"` e `tab_size = 2`.
- **Passos:**
  1. Navegar até a opção `[Salvar e Fechar]` e pressionar `Enter`.
  2. Inspecionar o arquivo gravado em disco.
- **Resultado Esperado:** O arquivo `config.toml` é gravado com sucesso sem arquivos temporários residuais, o conteúdo TOML contém `theme = "nord"` e `tab_size = 2`, o painel se fecha e a barra de status exibe: `"Configurações salvas em config.toml"`.
- **Covers:** Critério PRD: pressionar `Enter` na confirmação grava as configurações alteradas no arquivo `config.toml` de forma atômica e fecha o modal.

### TC15 — Coordenação com File Watcher Durante a Gravação
- **Alvo:** `internal/app/model.go` e `internal/config/watcher.go`
- **Tipo:** integration
- **Pré-condições:** `config.Watcher` ativo e monitorando o arquivo de configuração de teste.
- **Passos:**
  1. Executar o salvamento das configurações via painel.
  2. Inspecionar as chamadas de pausa e retomada do watcher.
- **Resultado Esperado:** O watcher é pausado antes da escrita atômica e retomado logo após a renomeação, prevenindo disparo espúrio de `ThemeReloadedMsg`.
- **Covers:** Cross-Feature: gravação de preferências persiste TOML e evita loops infinitos.

### TC16 — Tratamento de Erro de Permissão sem Travar a Sessão
- **Alvo:** `internal/app/model.go` e `internal/ui/settings.go`
- **Tipo:** integration
- **Pré-condições:** Arquivo de configuração em diretório protegido somente para leitura (`0555`).
- **Passos:**
  1. Tentar salvar as alterações pelo painel.
- **Resultado Esperado:** A tentativa de gravação falha sem gerar pânico no processo; o diálogo permanece aberto e exibe alerta visual em vermelho: `"Erro ao gravar config.toml: permissão negada"`.
- **Covers:** Critério PRD: em caso de erro de gravação em disco, exibe aviso de falha na interface sem travar a sessão.

### TC17 — Bloqueio Estrito de Propagação de Teclas para o Buffer
- **Alvo:** `internal/app/model.go` (`Update`)
- **Tipo:** integration
- **Pré-condições:** Painel de configurações aberto sobre um buffer de texto ativo.
- **Passos:**
  1. Pressionar teclas de edição normais do Vim (`i`, `x`, `d`, `dd`, `o`).
  2. Verificar o conteúdo do buffer de texto.
- **Resultado Esperado:** Nenhuma tecla enviada atinge o `buffer.Buffer` ou o `vim.Engine`; o texto do documento permanece completamente inalterado.
- **Covers:** Critério PRD / Cross-Feature: direciona temporariamente a captura de teclas exclusivamente ao modal.

### TC18 — Renderização Centralizada e Algoritmo de Overlay Blending
- **Alvo:** `internal/ui/settings.go` (`OverlayModal`)
- **Tipo:** integration
- **Pré-condições:** Buffer de 20 linhas preenchido com caracteres de teste em viewport 80x24. Modal gerado com dimensões 56x16.
- **Passos:**
  1. Executar `OverlayModal(backgroundLines, modalBox, 80, 24)`.
  2. Comparar as linhas mescladas resultantes.
- **Resultado Esperado:** As linhas centrais contêm o conteúdo da moldura do diálogo perfeitamente alinhadas, enquanto as linhas superiores, inferiores e margens laterais preservam os caracteres originais do documento sem quebras visuais.
- **Covers:** Cross-Feature: Viewport centraliza e renderiza a sobreposição mantendo o documento de fundo íntegro.

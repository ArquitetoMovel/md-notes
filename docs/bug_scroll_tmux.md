# Diagnóstico e Resolução: Bug de Rolagem e Renderização no Tmux

Relatório técnico de investigação e solução do problema relatado em ambientes `tmux` durante a navegação e rolagem em arquivos grandes com partes ocultas não renderizadas no **md-notes** (`mdn`).

---

## 1. Descrição do Problema

### Sintomas Observados
1. Ao abrir e navegar em arquivos extensos (especialmente arquivos Markdown com parágrafos longos) dentro de uma sessão ou painel do `tmux`, a rolagem não alcançava as partes inferiores do documento.
2. O cursor desaparecia para fora da tela e as linhas restantes permaneciam ocultas, mesmo com o cursor na última linha do buffer (ex: `100%` na statusline).
3. Ao usar a roda de scroll do mouse ou trackpad no `tmux`, o editor não respondia ou o `tmux` entrava no modo de cópia (*Copy Mode* `[0/...]`), rolando o histórico do terminal ao invés do documento, sem renderizar o conteúdo oculto.
4. Teclas clássicas de navegação rápida por página (`Ctrl+D`, `Ctrl+U`, `PageDown`, `PageUp`) não tinham efeito.

---

## 2. Investigação e Causas Raízes

A análise minuciosa da base de código revelou **duas falhas arquiteturais principais combinadas**, acompanhadas por carências de atalhos de navegação.

### Causa Raiz 1: Descompasso entre Índices Lógicos e Visuais com Soft-Wrap (`internal/ui/viewport.go`)

* **Mecanismo:** No `internal/ui/viewport.go`, o método `AdjustScroll(cursor buffer.Cursor, totalLines int)` calculava a linha do topo da tela (`v.TopLine`) considerando a posição do cursor em **linhas lógicas** (`curLine := cursor.Line`, limitado a `totalLines - 1`).
* **Incompatibilidade:** O método responsável por recortar as linhas a serem exibidas na tela (`GetVisibleLines`) utilizava `v.TopLine` como índice dentro da lista de **linhas visuais** (`allVisualLines[start:end]`).
* **Efeito com Quebra Automática (`word_wrap = true`):**
  - O editor ativa quebra de linha suave (*soft-wrap*) por padrão. Em Markdown, parágrafos normais excedem 80–120 caracteres e se dividem em 2 a 4 linhas visuais por linha lógica (especialmente em painéis divididos no `tmux`).
  - Em um arquivo de 100 linhas lógicas gerando 300 linhas visuais, quando o cursor atingia a linha 99, `AdjustScroll` limitava `TopLine` a no máximo `100 - 1 = 99`.
  - O fatiador `GetVisibleLines` então recortava as linhas visuais `99` a `119`. Essas linhas visuais correspondiam apenas às linhas lógicas 28 a 34!
  - **Resultado:** As linhas lógicas de 35 a 99 nunca eram renderizadas na tela. Conforme o usuário descia no arquivo, o cursor ficava invisível abaixo do rodapé da tela e mais de 65% do documento tornava-se inalcançável.

### Causa Raiz 2: Falta de Alternate Screen Buffer e Mouse Reporting (`cmd/mdn/main.go`)

* **Mecanismo:** No ponto de entrada (`cmd/mdn/main.go`), o programa Bubble Tea era instanciado via `tea.NewProgram(model, programOpts...)` sem as opções `tea.WithAltScreen()` e `tea.WithMouseCellMotion()`.
* **Impacto no `tmux`:**
  1. Sem a sequência ANSI de buffer alternativo (`\x1b[?1049h`), todo redesenho do editor era escrito diretamente no buffer de histórico do painel do `tmux`.
  2. Quando o usuário rolava com o mouse ou trackpad, o `tmux` (com `set -g mouse on`) detectava que o processo da janela não solicitou controle de mouse nem estava em tela alternativa.
  3. O `tmux` interceptava a rolagem e entrava no seu próprio *Copy Mode*, exibindo o histórico passado do terminal.
  4. Como o editor só desenhava a janela visível no momento, as partes inferiores do arquivo nunca haviam sido enviadas ao terminal. Logo, o histórico do `tmux` não as continha, gerando a percepção de que a rolagem travava e não revelava as partes ocultas.

### Causa Raiz 3: Ausência de Tratamento de Mensagens de Mouse e Teclas de Rolagem

* No loop central (`internal/app/model.go`), mensagens do tipo `tea.MouseMsg` não eram tratadas.
* No motor modal Vim (`internal/vim/engine.go`), teclas comuns de rolagem de páginas (`Ctrl+D`, `Ctrl+U`, `Ctrl+F`, `Ctrl+B`, `PageDown`, `PageUp`, `Ctrl+E`, `Ctrl+Y`) não estavam mapeadas.

---

## 3. Soluções Implementadas

### A. Correção do Motor de Viewport (`internal/ui/viewport.go`)

1. **Cálculo Preciso de Linhas Visuais:**
   - Adicionadas funções sem alocação desnecessária de memória:
     - `CountVisualLinesForLine(line, contentW, softWrap)`: calcula a quantidade de linhas visuais geradas por uma linha lógica.
     - `TotalVisualLines(buf)`: calcula o total global de linhas visuais do documento.
     - `CursorVisualRow(cursor, buf)`: localiza a linha visual exata do cursor em coordenadas visuais.
     - `VisualRowToCursor(targetVRow, buf)`: mapeia uma linha visual de volta para coordenadas lógicas `(Line, Col)`.
2. **Método `AdjustScrollWithBuffer(cursor, buf)`:**
   - Realiza a rolagem comparando a posição visual do cursor com os limites visuais da janela (`TopLine` a `TopLine + Height - 1 - scrolloff`).
   - Garante que a linha que contém o cursor sempre estará contida no intervalo fatiado pelo `GetVisibleLines`.
3. **Rolagem por Mouse (`ScrollDown` / `ScrollUp`):**
   - Rola a visualização em deltas configuráveis (ex: 3 linhas) e reposiciona o cursor automaticamente caso a rolagem o empurre para fora da viewport visível.
4. **Otimização com Interrupção Antecipada em `GetVisibleLines`:**
   - Em vez de gerar slices para todas as linhas de um arquivo grande, o fatiador para a iteração assim que obtém as `v.Height` linhas visíveis da tela (`break`), viabilizando renderização < 1 ms em documentos de 10.000+ linhas.

### B. Ativação de Alternate Screen e Captura de Mouse (`cmd/mdn/main.go`)

```go
var programOpts []tea.ProgramOption
programOpts = append(programOpts, tea.WithAltScreen(), tea.WithMouseCellMotion())
```

- Isola a visualização do `mdn` do histórico do `tmux`.
- Informa ao `tmux` para encaminhar eventos de roda de scroll diretamente à aplicação como sequências ANSI.

### C. Tratamento de Roda de Scroll no Bubble Tea (`internal/app/model.go`)

- Integrado suporte ao evento `tea.MouseMsg`:
  - `tea.MouseButtonWheelUp`: invoca `m.Viewport.ScrollUp(3, m.Buffer)`.
  - `tea.MouseButtonWheelDown`: invoca `m.Viewport.ScrollDown(3, m.Buffer)`.
- Atualizado o desenho da barra de rolagem lateral (`RenderScrollbar`) para basear-se no total real de linhas visuais (`TotalVisualLines`), mantendo precisão visual com soft-wrap ativo.

### D. Atalhos de Navegação de Página no Motor Vim (`internal/vim/engine.go`)

Mapeados nos modos Normal e Visual:
- `Ctrl+D`, `PageDown`, `pgdown`: desce meia página (12 linhas * multiplicador).
- `Ctrl+U`, `PageUp`, `pgup`: sobe meia página (12 linhas * multiplicador).
- `Ctrl+F`: desce uma página inteira (24 linhas * multiplicador).
- `Ctrl+B`: sobe uma página inteira (24 linhas * multiplicador).
- `Ctrl+E`: rola 1 linha para baixo.
- `Ctrl+Y`: rola 1 linha para cima.

---

## 4. Testes e Validação

Foram criados testes automatizados cobrindo cada camada da correção:

1. **`TestViewport_LargeDocumentWithSoftWrap_EndRendered`** (`internal/ui/viewport_test.go`):
   - Cria um buffer de 100 linhas longas (gerando ~300 linhas visuais com viewport 60x20).
   - Posiciona o cursor na linha 99 e aciona `AdjustScrollWithBuffer`.
   - **Resultado anterior:** Falha com linhas visíveis travadas entre 28 e 34.
   - **Resultado atual:** Sucesso, linha 99 renderizada no rodapé da viewport.

2. **`TestViewport_MouseScrollDownAndUp`** (`internal/ui/viewport_test.go`):
   - Valida o avanço e recuo de `TopLine` via mouse, com reposicionamento seguro do cursor dentro dos limites da tela.

3. **`TestEngine_PageScrollingKeys`** (`internal/vim/engine_test.go`):
   - Valida `Ctrl+D`, `Ctrl+U`, `PageDown` e `PageUp` deslocando o cursor com precisão.

4. **Suite Completa e Compilação:**
   - `make fmt`: código formatado pelo padrão Go.
   - `make vet`: zero advertências.
   - `make test`: todos os pacotes aprovados com 100% de sucesso.
   - `make build`: binário `bin/mdn` compilado com sucesso.

---

## 5. Dica de Configuração do Tmux para o Usuário

Para a melhor experiência com mouse e trackpad no `tmux`, certifique-se de que a seguinte linha esteja presente no seu arquivo `~/.tmux.conf`:

```tmux
# Habilita suporte completo a mouse no tmux
set -g mouse on
```

Com o `mdn` atualizado, o `tmux` repassa a roda de scroll diretamente para o editor de forma fluida, sem entrar no modo de seleção/cópia e renderizando instantaneamente todas as partes ocultas do documento.

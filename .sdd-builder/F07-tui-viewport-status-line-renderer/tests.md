# Especificação de Casos de Teste: F07. TUI Viewport & Status Line Renderer

## Sumário de Cobertura

| Tipo de Teste | Quantidade | Foco / Alvo |
|---|---|---|
| Unitário | 10 | Numeração de linhas, segmentos de status line, scrollbar, cursor ANSI e soft-wrap |
| Integração | 6 | Redimensionamento WindowSizeMsg, scrolloff, composição completa de tela e temas |
| Performance | 2 | Taxa de atualização constante a 60 FPS (< 16 ms por frame) em arquivos grandes |

---

## Matriz de Rastreabilidade

| Critério de Aceitação (PRD Seção 9) | Casos de Teste |
|---|---|
| A área de edição renderiza com precisão apenas as linhas visíveis na janela do terminal | TC01, TC02 |
| A coluna de números de linha suporta numeração absoluta e relativa configuráveis | TC03, TC04, TC05 |
| Linhas longas que ultrapassam a largura da janela sofrem soft-wrap visual sem alterar o buffer | TC06, TC07 |
| A barra de status exibe Modo, Arquivo, Status Dirty `[+]`, Cursor `Ln/Col`, Porcentagem e Encoding `UTF-8` | TC08, TC09, TC10 |
| O evento `tea.WindowSizeMsg` ajusta as dimensões do viewport e barra de status em tempo real | TC11, TC12 |
| Cross-Feature: Integração harmoniosa com Buffer, Tema, Vim, Markdown, Tabelas e Busca | TC13, TC14 |
| Full Scope: Scrollbar minimalista, cursores ANSI (bloco/barra/sublinhado) e 60 FPS | TC15, TC16, TC17, TC18 |

---

## Casos de Teste Unitários

### TC01 — Renderização Exclusiva de Linhas Visíveis
- **Alvo:** `internal/ui/viewport.go`
- **Tipo:** unit
- **Pré-condições:** Buffer com 100 linhas. Viewport com altura de 20 linhas e `TopLine = 10`.
- **Passos:**
  1. Processar a renderização do viewport.
  2. Inspecionar quais linhas foram desenhadas.
- **Resultado Esperado:** Apenas as linhas de índice 10 a 29 são processadas e renderizadas, descartando as 80 linhas fora da visão.
- **Cobre:** Critério PRD: renderiza com precisão apenas as linhas visíveis.

### TC02 — Ajuste de Margem de Rolagem Suave (*Scrolloff*)
- **Alvo:** `internal/ui/viewport.go` (`AdjustScroll`)
- **Tipo:** unit
- **Pré-condições:** `Scrolloff = 4`, altura do viewport = 20, cursor na linha 18 com `TopLine = 0`.
- **Passos:**
  1. Invocar `AdjustScroll`.
- **Resultado Esperado:** O `TopLine` é ajustado automaticamente para manter ao menos 4 linhas visíveis abaixo do cursor.
- **Cobre:** Navegação fluida com margem de scrolloff.

### TC03 — Numeração Absoluta de Linhas no Gutter
- **Alvo:** `internal/ui/gutter.go`
- **Tipo:** unit
- **Pré-condições:** `relative = false`, cursor na linha 5.
- **Passos:**
  1. Renderizar linhas 1 a 10 do gutter.
- **Resultado Esperado:** As linhas exibem `1, 2, 3, 4, 5, 6, 7, 8, 9, 10` alinhadas à direita com padding.
- **Cobre:** Critério PRD: suporte a numeração absoluta.

### TC04 — Numeração Relativa de Linhas no Gutter
- **Alvo:** `internal/ui/gutter.go`
- **Tipo:** unit
- **Pré-condições:** `relative = true`, cursor na linha 5 (1-based).
- **Passos:**
  1. Renderizar linhas 3 a 7 do gutter.
- **Resultado Esperado:** A linha 5 exibe o número absoluto `5`, enquanto as linhas 3, 4, 6 e 7 exibem respectivamente as distâncias relativas `2, 1, 1, 2`.
- **Cobre:** Critério PRD: suporte a numeração relativa configurável.

### TC05 — Largura Adaptativa da Coluna de Gutter
- **Alvo:** `internal/ui/gutter.go`
- **Tipo:** unit
- **Pré-condições:** Buffer com 12.500 linhas.
- **Passos:**
  1. Calcular a largura necessária da coluna.
- **Resultado Esperado:** A largura calculada acomoda 5 dígitos mais os espaços de padding (`5 + 2 = 7 colunas`).
- **Cobre:** Formatação de gutter sem truncar dígitos.

### TC06 — Quebra de Linha Suave (*Soft-Wrap*) sem Mutação do Buffer
- **Alvo:** `internal/ui/viewport.go`
- **Tipo:** unit
- **Pré-condições:** Linha única no buffer com 150 caracteres. Largura disponível de conteúdo de 60 colunas.
- **Passos:**
  1. Processar a linha no motor de viewport com soft-wrap ativo.
  2. Verificar o número de linhas visuais geradas e o conteúdo original do buffer.
- **Resultado Esperado:** A linha quebra em 3 linhas visuais na tela enquanto o buffer original mantém uma única linha sem quebras `\n`.
- **Cobre:** Critério PRD: soft-wrap visual sem alterar o buffer.

### TC07 — Mapeamento de Coordenadas com Soft-Wrap
- **Alvo:** `internal/ui/viewport.go`
- **Tipo:** unit
- **Pré-condições:** Linha longa com quebra visual na coluna 60. Cursor posicionado na coluna lógica 80.
- **Passos:**
  1. Consultar a posição visual de tela `(ScreenY, ScreenX)`.
- **Resultado Esperado:** O cursor é mapeado para a segunda linha visual (`ScreenY = 1`) na coluna `20` (`ScreenX = 20`).
- **Cobre:** Sincronização de posição de cursor em linhas com soft-wrap.

### TC08 — Segmentos Essenciais da Barra de Status
- **Alvo:** `internal/ui/statusline.go`
- **Tipo:** unit
- **Pré-condições:** Estado: Modo Normal, Arquivo `"docs/PRD.md"`, `IsDirty = false`, `Ln 45, Col 12`, `UTF-8 | LF`.
- **Passos:**
  1. Renderizar a barra de status.
- **Resultado Esperado:** Contém blocos formatados com `NORMAL`, `docs/PRD.md`, `Ln 45, Col 12`, `UTF-8` e `LF`.
- **Cobre:** Critério PRD: exibição de Modo, Arquivo, Cursor e Encoding.

### TC09 — Indicador Visual de Modificação `[+]`
- **Alvo:** `internal/ui/statusline.go`
- **Tipo:** unit
- **Pré-condições:** `IsDirty = true`.
- **Passos:**
  1. Renderizar a barra de status.
- **Resultado Esperado:** O nome do arquivo é acompanhado pelo indicador `[+]` em destaque visual.
- **Cobre:** Critério PRD: indicador de status Dirty `[+]`.

### TC10 — Cálculo de Progresso Percentual (`Top`, `50%`, `Bot`)
- **Alvo:** `internal/ui/statusline.go`
- **Tipo:** unit
- **Pré-condições:** Buffer de 100 linhas testado na linha 1, na linha 50 e na linha 100.
- **Passos:**
  1. Renderizar indicador de progresso para cada caso.
- **Resultado Esperado:** Linha 1 exibe `"Top"`; linha 50 exibe `"50%"`; linha 100 exibe `"Bot"`.
- **Cobre:** Critério PRD: porcentagem de navegação na barra de status.

---

## Casos de Teste de Integração

### TC11 — Redimensionamento Dinâmico de Janela (`tea.WindowSizeMsg`)
- **Alvo:** `internal/app/model.go` & `internal/ui/viewport.go`
- **Tipo:** integration
- **Pré-condições:** Aplicação em execução com 80x24.
- **Passos:**
  1. Enviar mensagem `tea.WindowSizeMsg{Width: 120, Height: 40}`.
  2. Verificar dimensões do viewport e barra de status.
- **Resultado Esperado:** As dimensões internas são atualizadas instantaneamente e a renderização subsequente ocupa a largura de 120 colunas e altura de 40 linhas.
- **Cobre:** Critério PRD: redimensionamento dinâmico em tempo real via WindowSizeMsg.

### TC12 — Exibição de Indicador de Busca Ativa na Barra de Status
- **Alvo:** `internal/ui/statusline.go` & `internal/search`
- **Tipo:** integration
- **Pré-condições:** Busca ativa com ocorrência 3 de 15 selecionada.
- **Passos:**
  1. Renderizar a barra de status.
- **Resultado Esperado:** A barra exibe o bloco `[3/15]` com destaque de cor do tema.
- **Cobre:** Integração com Search Engine (F06).

### TC13 — Composição Visual Completa (Gutter + Viewport + Status Line)
- **Alvo:** `internal/app/model.go` (`View`)
- **Tipo:** integration
- **Pré-condições:** Documento com títulos Markdown (F04), tabelas (F05) e modo Normal (F03).
- **Passos:**
  1. Invocar `View()`.
- **Resultado Esperado:** A string final gerada contém a coluna de números, o conteúdo enriquecido e alinhado, a scrollbar na margem direita e a barra de status na base sem linhas vazias excedentes.
- **Cobre:** Critério Cross-Feature: orquestração visual de todos os módulos.

### TC14 — Adaptação Dinâmica a Troca de Tema (F02)
- **Alvo:** `internal/ui/statusline.go` & `internal/theme`
- **Tipo:** integration
- **Pré-condições:** Alternar tema de `nord` para `dracula`.
- **Passos:**
  1. Renderizar o viewport e a barra de status.
- **Resultado Esperado:** Todas as cores de fundo, texto e bordas adotam instantaneamente a paleta Dracula.
- **Cobre:** Critério Cross-Feature: consumo de CompiledTheme.

---

## Casos de Teste Full Scope

### TC15 — Renderização de Barra de Rolagem Minimalista (*Scrollbar*)
- **Alvo:** `internal/ui/scrollbar.go`
- **Tipo:** unit / integration
- **Pré-condições:** Documento de 200 linhas em janela de 20 linhas. Posição atual no meio do arquivo.
- **Passos:**
  1. Gerar os caracteres da barra de rolagem para as 20 linhas.
- **Resultado Esperado:** As linhas centrais recebem o caractere de thumb (`█`) enquanto as extremidades recebem trilho sutil (`│`).
- **Cobre:** Full Scope PRD: barra de rolagem visual minimalista.

### TC16 — Formatos de Cursor ANSI por Modo
- **Alvo:** `internal/ui/cursor.go`
- **Tipo:** unit
- **Pré-condições:** Alternância entre os modos Normal, Insert e Visual.
- **Passos:**
  1. Consultar a sequência ANSI emitida para cada modo.
- **Resultado Esperado:** Modo Normal emite `\x1b[2 q` (bloco); Modo Insert emite `\x1b[6 q` (barra); Modo Visual emite `\x1b[4 q` (sublinhado).
- **Cobre:** Full Scope PRD: modos de cursor ANSI nativos.

### TC17 — Desempenho de Renderização Contínua a 60 FPS (< 16 ms por frame)
- **Alvo:** `internal/app/model.go` (`View`)
- **Tipo:** performance / benchmark
- **Pré-condições:** Janela de terminal 160x50 com documento de 50.000 linhas contendo Markdown complexo, tabelas e busca ativa.
- **Passos:**
  1. Executar 500 ciclos consecutivos de `View()`.
- **Resultado Esperado:** Tempo médio de renderização por frame inferior a 5 milissegundos (muito abaixo do limite de 16,6 ms de 60 FPS).
- **Cobre:** Sucesso PRD: taxa de atualização interativa a 60 FPS sem latência perceptível.

### TC18 — Restauração de Cursor e Terminal no Encerramento
- **Alvo:** `internal/ui/cursor.go` & `internal/app`
- **Tipo:** unit / integration
- **Pré-condições:** Fechamento do editor via `:q` ou `Ctrl+c`.
- **Passos:**
  1. Validar a sequência de saída emitida.
- **Resultado Esperado:** É emitida a sequência `\x1b[0 q` restaurando a configuração original de cursor do terminal do usuário.
- **Cobre:** Integridade do terminal do usuário após a execução.

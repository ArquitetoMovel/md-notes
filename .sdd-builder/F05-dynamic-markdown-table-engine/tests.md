# Especificação de Casos de Teste: F05. Dynamic Markdown Table Engine

## Sumário de Cobertura

| Tipo de Teste | Quantidade | Foco / Alvo |
|---|---|---|
| Unitário | 9 | Parser de delimitadores, cálculo de largura Unicode/emojis, padding e alinhamentos |
| Integração | 6 | Navegação com Tab/Shift+Tab, criação de linhas, inserção de colunas e bordas Unicode |
| E2E / TUI | 2 | Edição interativa de tabelas no modelo Bubble Tea |

---

## Matriz de Rastreabilidade

| Critério de Aceitação (PRD Seção 9) | Casos de Teste |
|---|---|
| Tabelas Markdown são detectadas e alinhadas dinamicamente com base no conteúdo de cada coluna | TC01, TC02 |
| Alinhamentos à esquerda (`:---`), centralizado (`:---:`) e à direita (`---:`) são respeitados na renderização | TC03, TC04, TC05 |
| A tecla `Tab` dentro de uma tabela avança para a próxima célula e `Shift+Tab` retrocede para a anterior | TC06, TC07 |
| Ao pressionar `Tab` na última célula da tabela, uma nova linha é criada automaticamente com os pipes correspondentes | TC08 |
| A largura de caracteres multibyte e emojis é calculada corretamente no alinhamento das bordas | TC09, TC10 |
| Cross-Feature: Integração de coordenadas de cursor e renderização no Viewport (F07) | TC11, TC12 |
| Full Scope: Bordas estéticas Unicode, inserção/remoção de colunas e Markdown puro no buffer | TC13, TC14, TC15, TC16, TC17 |

---

## Casos de Teste Unitários

### TC01 — Detecção Automática de Tabela Markdown GFM
- **Alvo:** `internal/table/parser.go` (`DetectTable`)
- **Tipo:** unit
- **Pré-condições:** Conjunto de linhas contendo cabeçalho `| A | B |` e separador `|---|---|`.
- **Passos:**
  1. Invocar `DetectTable` com o cursor posicionado na linha do cabeçalho.
  2. Verificar retorno booleano e estrutura `Table`.
- **Resultado Esperado:** Detecta o início e o fim do bloco de tabela e extrai corretamente as células do cabeçalho e dados.
- **Cobre:** Critério PRD: detecção automática de tabelas Markdown.

### TC02 — Rejeição de Linhas de Texto com Pipe Isolado
- **Alvo:** `internal/table/parser.go`
- **Tipo:** unit
- **Pré-condições:** Linha de texto comum contendo `"comando1 | comando2"` sem linha de separador subsequente.
- **Passos:**
  1. Invocar `DetectTable`.
- **Resultado Esperado:** Retorna `false` indicando que a linha não faz parte de uma tabela estruturada.
- **Cobre:** Prevenção de falsos positivos na detecção de tabelas.

### TC03 — Alinhamento de Colunas à Esquerda (`:---` ou `---`)
- **Alvo:** `internal/table/formatter.go`
- **Tipo:** unit
- **Pré-condições:** Tabela com separador `|:---|---|` e células de comprimentos variados.
- **Passos:**
  1. Formatar a tabela via `FormatMarkdownTable`.
  2. Verificar distribuição de espaços nas células.
- **Resultado Esperado:** O texto de cada célula fica colado à esquerda com espaços de padding preenchendo o lado direito até a largura máxima.
- **Cobre:** Critério PRD: alinhamento à esquerda respeitado.

### TC04 — Alinhamento de Colunas Centralizado (`:---:`)
- **Alvo:** `internal/table/formatter.go`
- **Tipo:** unit
- **Pré-condições:** Tabela com separador `|:---:|` e coluna de largura 10 com texto de 4 caracteres (`"Item"`).
- **Passos:**
  1. Formatar a tabela.
  2. Verificar padding à esquerda e à direita do texto.
- **Resultado Esperado:** O texto recebe 3 espaços à esquerda e 3 espaços à direita, ficando perfeitamente centralizado.
- **Cobre:** Critério PRD: alinhamento centralizado respeitado.

### TC05 — Alinhamento de Colunas à Direita (`---:`)
- **Alvo:** `internal/table/formatter.go`
- **Tipo:** unit
- **Pré-condições:** Tabela com separador `|---:|` contendo valores numéricos (`"123"`, `"45.60"`).
- **Passos:**
  1. Formatar a tabela.
- **Resultado Esperado:** O texto é encostado à direita da célula, com espaços de padding posicionados à esquerda.
- **Cobre:** Critério PRD: alinhamento à direita respeitado.

### TC06 — Cálculo de Largura com Caracteres Acentuados e Multibyte
- **Alvo:** `internal/table/formatter.go`
- **Tipo:** unit
- **Pré-condições:** Célula contendo `"Atenção e Configuração"`.
- **Passos:**
  1. Calcular a largura visual da célula.
- **Resultado Esperado:** A largura computada é exatamente 22 colunas visuais, sem quebras por bytes UTF-8 extras.
- **Cobre:** Critério PRD: cálculo de largura de caracteres multibyte.

### TC07 — Cálculo de Largura com Emojis de Largura Dupla (Width = 2)
- **Alvo:** `internal/table/formatter.go`
- **Tipo:** unit
- **Pré-condições:** Célula contendo `"Status: 🚀 Concluído"`.
- **Passos:**
  1. Calcular largura da célula com `go-runewidth`.
- **Resultado Esperado:** O emoji `🚀` é computado como 2 colunas visuais, totalizando 20 colunas e mantendo o alinhamento das barras perfeitamente reto.
- **Cobre:** Critério PRD: cálculo correto de emojis no alinhamento de bordas.

### TC08 — Navegação para a Próxima Célula com `Tab`
- **Alvo:** `internal/table/navigation.go` (`NextCell`)
- **Tipo:** unit
- **Pré-condições:** Tabela de 3 colunas com cursor na coluna 0 da linha 1.
- **Passos:**
  1. Executar `NextCell`.
- **Resultado Esperado:** Retorna `nextRow = 1`, `nextCol = 1`, `isNewRow = false`.
- **Cobre:** Critério PRD: tecla Tab avança para a próxima célula.

### TC09 — Navegação para a Célula Anterior com `Shift+Tab`
- **Alvo:** `internal/table/navigation.go` (`PrevCell`)
- **Tipo:** unit
- **Pré-condições:** Tabela com cursor na coluna 1 da linha 1.
- **Passos:**
  1. Executar `PrevCell`.
- **Resultado Esperado:** Retorna `prevRow = 1`, `prevCol = 0`.
- **Cobre:** Critério PRD: tecla Shift+Tab retrocede para a célula anterior.

---

## Casos de Teste de Integração

### TC10 — Criação Automática de Nova Linha com `Tab` na Última Célula
- **Alvo:** `internal/table/navigation.go` & `internal/buffer`
- **Tipo:** integration
- **Pré-condições:** Tabela de 3 colunas com cursor na última célula da última linha.
- **Passos:**
  1. Pressionar `Tab`.
  2. Verificar as linhas adicionadas ao buffer.
- **Resultado Esperado:** Uma nova linha `|   |   |   |` é inserida na tabela e o cursor é posicionado no início da primeira célula da nova linha.
- **Cobre:** Critério PRD: Tab na última célula cria automaticamente nova linha.

### TC11 — Preservação Estrita de Sintaxe Markdown Pura no Buffer
- **Alvo:** `internal/table/formatter.go`
- **Tipo:** integration
- **Pré-condições:** Tabela editada e reformatada.
- **Passos:**
  1. Inspecionar as linhas brutas persistidas no buffer (`Buffer.GetRawText()`).
- **Resultado Esperado:** As linhas contêm apenas caracteres ASCII padrão (`|`, `-`, espaços) sem caracteres Unicode de desenho de caixa gravados no arquivo.
- **Cobre:** Capacidade PRD: preservação de sintaxe Markdown pura no disco.

### TC12 — Renderização de Bordas Estéticas Unicode de Caixa (`┌─┬─┐`)
- **Alvo:** `internal/table/renderer.go`
- **Tipo:** integration
- **Pré-condições:** Tabela com 2 colunas e configuração `editor.table_borders = "unicode"`.
- **Passos:**
  1. Renderizar tabela via `RenderUnicode(theme)`.
- **Resultado Esperado:** Gera molduras contendo `┌`, `┬`, `┐` no topo, `├`, `┼`, `┤` entre cabeçalho e dados, e `└`, `┴`, `┘` na base.
- **Cobre:** Full Scope PRD: bordas estéticas Unicode de desenho de caixa.

### TC13 — Inserção Rápida de Coluna à Direita (`Alt+Shift+L`)
- **Alvo:** `internal/table/navigation.go`
- **Tipo:** integration
- **Pré-condições:** Tabela com 2 colunas. Cursor na coluna 0.
- **Passos:**
  1. Disparar `InsertColumn(atRight: true)`.
  2. Inspecionar todas as linhas da tabela.
- **Resultado Esperado:** Uma nova coluna vazia é inserida na posição 1 em todas as linhas (cabeçalho, delimitador e dados).
- **Cobre:** Full Scope PRD: atalhos de inserção de colunas.

### TC14 — Inserção Rápida de Coluna à Esquerda (`Alt+Shift+H`)
- **Alvo:** `internal/table/navigation.go`
- **Tipo:** integration
- **Pré-condições:** Tabela com 2 colunas. Cursor na coluna 1.
- **Passos:**
  1. Disparar `InsertColumn(atRight: false)`.
- **Resultado Esperado:** Uma nova coluna vazia é inserida antes da coluna atual em todas as linhas da tabela.
- **Cobre:** Full Scope PRD: inserção de coluna à esquerda.

### TC15 — Deleção de Coluna sob o Cursor (`Alt+Shift+X`)
- **Alvo:** `internal/table/navigation.go`
- **Tipo:** integration
- **Pré-condições:** Tabela com 3 colunas. Cursor na coluna 1.
- **Passos:**
  1. Disparar `DeleteColumn(colIdx: 1)`.
- **Resultado Esperado:** A coluna intermediária é removida de todas as linhas e a tabela passa a ter 2 colunas com alinhamento recalculado.
- **Cobre:** Full Scope PRD: deleção de colunas.

### TC16 — Formatação de Sintaxe Inline Dentro de Células de Tabela
- **Alvo:** `internal/table/renderer.go` & `internal/markdown`
- **Tipo:** integration
- **Pré-condições:** Célula contendo `"Texto em **negrito** e `código`"`.
- **Passos:**
  1. Renderizar a tabela.
- **Resultado Esperado:** Os spans inline são coloridos e estilizados mantendo a largura visual calculada da célula perfeitamente alinhada.
- **Cobre:** Full Scope PRD: realce inline dentro de células.

### TC17 — Estabilidade de Coordenadas de Cursor ao Reformatar
- **Alvo:** `internal/table/formatter.go` & `internal/buffer`
- **Tipo:** integration
- **Pré-condições:** Cursor digitando na célula `"Engenheira"` (coluna 3 do texto interno).
- **Passos:**
  1. Reformatar a linha da tabela.
- **Resultado Esperado:** O cursor continua apontando exatamente para o mesmo caractere do texto sem pular para fora da célula.
- **Cobre:** Critério Cross-Feature: preservação das coordenadas de cursor do Vim Engine.

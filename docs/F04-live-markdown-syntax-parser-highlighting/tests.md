# Especificação de Casos de Teste: F04. Live Markdown Syntax Parser & Highlighting

## Sumário de Cobertura

| Tipo de Teste | Quantidade | Foco / Alvo |
|---|---|---|
| Unitário | 10 | Spans inline, marcadores atenuados, títulos H1-H6, listas, citações e Chroma |
| Integração | 6 | Parsing de buffer multiline, formatações aninhadas, temas e links |
| Performance | 2 | Latência de parsing por linha (< 1ms) e renderização contínua a 60 FPS |

---

## Matriz de Rastreabilidade

| Critério de Aceitação (PRD Seção 9) | Casos de Teste |
|---|---|
| Linhas iniciadas com `#` até `######` são destacadas com cores hierárquicas e peso negrito em tempo real | TC01, TC02 |
| Os marcadores de formatação (`#`, `**`, `*`, `_`, `~~`, `` ` ``) são exibidos com estilo atenuado (*dimmed*) | TC03, TC04 |
| Textos delimitados por `**` recebem estilo negrito e textos com `*` ou `_` recebem estilo itálico | TC05, TC06 |
| Blocos de código delimitados por ``` recebem fundo destacado e recuo visual | TC07, TC08 |
| Listas com `- `, `* `, `+ `, `1. ` e citações com `> ` são formatadas com cores e indentação adequadas | TC09, TC10, TC11 |
| Cross-Feature: Linhas tokenizadas são fornecidas para F05 (Tables) e F07 (Viewport) | TC12, TC13 |
| Full Scope: Syntax highlighting com Chroma, links/URLs, formatações aninhadas e performance | TC14, TC15, TC16, TC17, TC18 |

---

## Casos de Teste Unitários

### TC01 — Destaque Hierárquico de Títulos (H1 a H6)
- **Alvo:** `internal/markdown/parser.go`
- **Tipo:** unit
- **Pré-condições:** Parser instanciado com tema ativo.
- **Passos:**
  1. Processar linhas contendo `"# Título 1"`, `"## Título 2"`, ate `"###### Título 6"`.
  2. Verificar `IsHeading`, `HeadingLvl` e estilos dos spans de cada linha.
- **Resultado Esperado:** Cada linha gera spans com `IsHeading == true` e `HeadingLvl` correspondente (1 a 6); o texto do título recebe estilo em negrito e cor hierárquica do nível correspondente no tema.
- **Cobre:** Critério PRD: títulos destacados com cores hierárquicas e peso negrito.

### TC02 — Atenuação de Marcadores de Título (`#`)
- **Alvo:** `internal/markdown/parser.go`
- **Tipo:** unit
- **Pré-condições:** Linha `"# Introdução ao Projeto"`.
- **Passos:**
  1. Analisar os spans gerados para a linha.
- **Resultado Esperado:** O primeiro span contém `"# "` com o estilo `theme.Muted`, e o segundo span contém `"Introdução ao Projeto"` com estilo `theme.H1` em negrito.
- **Cobre:** Critério PRD: marcadores `#` exibidos em tom atenuado (*dimmed*).

### TC03 — Atenuação de Marcadores Inline (`**`, `*`, `_`, `~~`, `` ` ``)
- **Alvo:** `internal/markdown/lexer.go`
- **Tipo:** unit
- **Pré-condições:** Linha com `**negrito**`, `*itálico*`, `~~tachado~~` e `` `código` ``.
- **Passos:**
  1. Tokenizar a linha com o lexer.
  2. Inspecionar o tipo e o estilo de cada `StyledSpan`.
- **Resultado Esperado:** Os spans correspondentes aos delimitadores (`**`, `*`, `~~`, `` ` ``) possuem `Type == TokenMarker` e estilo `theme.Muted`.
- **Cobre:** Critério PRD: marcadores sintáticos inline mantidos com estilo atenuado.

### TC04 — Formatação de Código Inline com Fundo Diferenciado
- **Alvo:** `internal/markdown/lexer.go`
- **Tipo:** unit
- **Pré-condições:** Linha contendo `"execute o comando `go test` para validar"`.
- **Passos:**
  1. Tokenizar a linha.
  2. Inspecionar o span contendo `"go test"`.
- **Resultado Esperado:** O span possui `Type == TokenCodeInline` com estilo `theme.CodeInline` (fundo sutil e cor de destaque).
- **Cobre:** Critério PRD: código inline renderizado com fundo e cor de destaque.

### TC05 — Formatação de Texto em Negrito (`**texto**`)
- **Alvo:** `internal/markdown/lexer.go`
- **Tipo:** unit
- **Pré-condições:** Linha contendo `"Este é um texto **importante e em negrito** aqui"`.
- **Passos:**
  1. Tokenizar a linha.
  2. Verificar o span intermediário.
- **Resultado Esperado:** O texto `"importante e em negrito"` recebe `Type == TokenBold` e estilo `theme.Bold`.
- **Cobre:** Critério PRD: textos delimitados por `**` recebem estilo negrito.

### TC06 — Formatação de Texto em Itálico (`*texto*` e `_texto_`)
- **Alvo:** `internal/markdown/lexer.go`
- **Tipo:** unit
- **Pré-condições:** Linha contendo `"*ênfase um* e _ênfase dois_"`.
- **Passos:**
  1. Tokenizar a linha.
  2. Verificar os spans intermediários.
- **Resultado Esperado:** Ambos os trechos recebem `Type == TokenItalic` e estilo `theme.Italic`.
- **Cobre:** Critério PRD: textos delimitados por `*` ou `_` recebem estilo itálico.

### TC07 — Abertura e Fechamento de Blocos de Código (```)
- **Alvo:** `internal/markdown/parser.go`
- **Tipo:** unit
- **Pré-condições:** Três linhas: ````go````, `"package main"`, ```` ````.
- **Passos:**
  1. Processar as três linhas em sequência mantendo o contexto de bloco.
- **Resultado Esperado:** A linha 1 ativa o modo bloco de código; a linha 2 é classificada com `IsCodeBlock == true`; a linha 3 encerra o bloco de código.
- **Cobre:** Critério PRD: blocos de código delimitados por crases triplas.

### TC08 — Fundo Visual e Recuo em Linhas de Código
- **Alvo:** `internal/markdown/parser.go`
- **Tipo:** unit
- **Pré-condições:** Linha intermediária de bloco de código `"fmt.Println(123)"`.
- **Passos:**
  1. Processar a linha com contexto de bloco ativo.
- **Resultado Esperado:** A linha recebe estilo `theme.CodeBlock` com cor de fundo distinta e recuo visual.
- **Cobre:** Critério PRD: linhas de bloco de código recebem fundo destacado e recuo.

### TC09 — Listas Não Ordenadas (`- `, `* `, `+ `)
- **Alvo:** `internal/markdown/parser.go`
- **Tipo:** unit
- **Pré-condições:** Linhas com `"- Item A"`, `"* Item B"`, `"+ Item C"`.
- **Passos:**
  1. Processar as linhas.
- **Resultado Esperado:** Os marcadores de lista são destacados com estilo de marcador e o texto do item mantém indentação alinhada.
- **Cobre:** Critério PRD: listas não ordenadas com cores e indentação adequadas.

### TC10 — Listas Ordenadas (`1. `, `2. `)
- **Alvo:** `internal/markdown/parser.go`
- **Tipo:** unit
- **Pré-condições:** Linhas com `"1. Primeiro passo"`, `"2. Segundo passo"`.
- **Passos:**
  1. Processar as linhas.
- **Resultado Esperado:** Os números e pontos são identificados como prefixo numérico de lista e formatados com cor secundária do tema.
- **Cobre:** Critério PRD: listas ordenadas formatadas com hierarquia.

### TC11 — Citações em Bloco (`> `) com Barra Decorativa
- **Alvo:** `internal/markdown/parser.go`
- **Tipo:** unit
- **Pré-condições:** Linha `"> Esta é uma citação importante"`.
- **Passos:**
  1. Processar a linha.
- **Resultado Esperado:** O prefixo `>` é substituído visualmente por uma barra vertical decorativa estilizada (`│ `) com cor do tema, e o texto recebe estilo itálico suave.
- **Cobre:** Critério PRD: citações com `> ` formatadas com barra decorativa.

---

## Casos de Teste de Integração

### TC12 — Fornecimento de `TokenizedLine` para Outros Módulos
- **Alvo:** `internal/markdown/token.go`
- **Tipo:** integration
- **Pré-condições:** Parser processando um buffer de teste.
- **Passos:**
  1. Obter `[]TokenizedLine` geradas pelo parser.
  2. Validar que cada linha expõe `Spans`, `Original`, `Render()` e `PlainText()`.
- **Resultado Esperado:** A estrutura de dados fornece acesso $O(1)$ aos spans e textos renderizados sem perda da linha original do buffer.
- **Cobre:** Critério Cross-Feature: linhas tokenizadas consumidas por F05 e F07.

### TC13 — Renderização ANSI Consistente com Temas Lipgloss (F02)
- **Alvo:** `internal/markdown/parser.go`
- **Tipo:** integration
- **Pré-condições:** Parser inicializado com tema `dracula` e depois com tema `nord`.
- **Passos:**
  1. Processar `"# Título"` em ambos os temas.
  2. Inspecionar as sequências ANSI geradas por `Render()`.
- **Resultado Esperado:** As cores das sequências ANSI mudam conforme a paleta ativa do tema sem erros de compilação.
- **Cobre:** Critério Cross-Feature: estilos de Theme Engine aplicados uniformemente.

---

## Casos de Teste Full Scope

### TC14 — Syntax Highlighting Chroma em Código Go e Python
- **Alvo:** `internal/markdown/highlighter.go`
- **Tipo:** unit / integration
- **Pré-condições:** Bloco de código Go: ````go\nfunc main() {\n\treturn\n}\n````.
- **Passos:**
  1. Processar as linhas com Chroma.
  2. Verificar classificação de tokens (`func` como keyword, `main` como identificador).
- **Resultado Esperado:** As palavras-chave recebem cores semânticas Truecolor distintas dos identificadores e delimitadores.
- **Cobre:** Full Scope PRD: syntax highlighting completo em blocos de código.

### TC15 — Fallback para Blocos de Código sem Linguagem Reconhecida
- **Alvo:** `internal/markdown/highlighter.go`
- **Tipo:** integration
- **Pré-condições:** Bloco ````linguagem_desconhecida_123\ntexto simples\n````.
- **Passos:**
  1. Processar com o highlighter.
- **Resultado Esperado:** O código é renderizado normalmente com estilo base de bloco de código sem erros ou interrupções.
- **Cobre:** Full Scope PRD: fallback seguro para linguagens não reconhecidas.

### TC16 — Detecção e Realce de Links e URLs (`[texto](url)`)
- **Alvo:** `internal/markdown/lexer.go`
- **Tipo:** unit
- **Pré-condições:** Linha `"Acesse o site [Documentação](https://golang.org) para mais detalhes."`
- **Passos:**
  1. Tokenizar a linha.
  2. Inspecionar spans gerados.
- **Resultado Esperado:** `"Documentação"` recebe estilo de link destacado, enquanto `[`, `]`, `(` e `https://golang.org)` recebem estilo atenuado `Muted`.
- **Cobre:** Full Scope PRD: suporte a links e URLs clicáveis no terminal.

### TC17 — Formatações Aninhadas Complexas
- **Alvo:** `internal/markdown/lexer.go`
- **Tipo:** unit
- **Pré-condições:** Linha contendo `"**negrito com `código inline` e *itálico* aninhado**"`.
- **Passos:**
  1. Tokenizar a linha.
- **Resultado Esperado:** O lexer processa os estilos de forma cumulativa, garantindo que o código inline mantenha seu fundo diferenciado dentro do bloco de negrito.
- **Cobre:** Capacidade PRD: suporte a formatações aninhadas.

### TC18 — Performance de Parsing Incremental (< 1 ms por linha)
- **Alvo:** `internal/markdown/parser.go`
- **Tipo:** performance / benchmark
- **Pré-condições:** Documento com 1.000 linhas contendo títulos, negrito, itálico, listas, tabelas e blocos de código.
- **Passos:**
  1. Medir o tempo de parsing de todo o documento em 100 iterações.
- **Resultado Esperado:** Tempo médio de processamento inferior a 0,5 ms por linha, assegurando taxa de atualização estável a 60 FPS.
- **Cobre:** Sucesso PRD: taxa de renderização interativa < 16 ms (60 FPS).

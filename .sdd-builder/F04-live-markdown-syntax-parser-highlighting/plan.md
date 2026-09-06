# Plano de Implementação: F04. Live Markdown Syntax Parser & Highlighting

**Pré-requisitos:**
- Go 1.22 ou superior instalado no ambiente de desenvolvimento
- Dependências Go declaradas em `go.mod`:
  - `github.com/charmbracelet/lipgloss` (v0.9+)
  - `github.com/alecthomas/chroma/v2` (v2.14+)
- Módulos `internal/buffer` (F01) e `internal/theme` (F02) implementados ou estruturados

---

### Fase 1: Definição de Tipos de Token e Estruturas de Spans Estilizados

**1. Tipos de Tokens e Estrutura de Spans** - Criar em `internal/markdown/token.go` as enumerações de tokens semânticos e as structs `StyledSpan` e `TokenizedLine` para armazenar segmentos de texto com estilos associados.

**2. Renderização de Linhas Tokenizadas** - Implementar métodos de concatenação e renderização de spans em texto estilizado aplicando sequências de escape ANSI compatíveis com o terminal.

---

### Fase 2: Lexer Inline e Destaque de Marcadores Atenuados

**3. Scanner Linear de Ênfases Inline** - Desenvolver em `internal/markdown/lexer.go` o analisador de texto inline para identificar negrito, itálico, tachado e código inline, aplicando estilo atenuado aos marcadores sintáticos.

**4. Suporte a Formatações Aninhadas** - Implementar no lexer a capacidade de processar combinações aninhadas de formatação em uma mesma sequência de texto sem quebra de integridade.

**5. Detecção e Destaque de Links e URLs** - Adicionar a identificação de referências a links Markdown e URLs diretas com formatação visual destacada e marcadores em tom discreto.

---

### Fase 3: Parser Estrutural de Blocos e Títulos

**6. Hierarquia de Títulos Markdown** - Construir em `internal/markdown/parser.go` a detecção de títulos de nível 1 a 6 com cores hierárquicas, negrito e caracteres de cerquilha atenuados.

**7. Blocos de Citação com Barra Decorativa** - Implementar o parser para linhas de citação iniciadas com sinal de maior, aplicando recuo e caracteres verticais estilizados.

**8. Listas e Checkboxes** - Desenvolver a formatação de itens de lista numerada, marcadores de lista não ordenada e caixas de seleção com símbolos visuais adequados.

**9. Contexto de Blocos de Código Multiline** - Criar o rastreador de estado para gerenciar linhas intermediárias contidas entre delimitadores de crases triplas.

---

### Fase 4: Syntax Highlighting com Chroma e Otimização de Performance

**10. Integração com o Motor Chroma** - Implementar em `internal/markdown/highlighter.go` o adaptador Chroma para coloração de sintaxe em múltiplas linguagens de programação com cache de lexers.

**11. Fundo Visual e Formatação de Blocos de Código** - Aplicar cores de fundo distintas e recuo visual uniforme para linhas de código dentro do editor.

**12. Otimização de Parsing Incremental** - Refinar o processamento de linhas para garantir latência de análise inferior a 1 milissegundo por linha em documentos extensos.

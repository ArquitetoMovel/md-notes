# Plano de Implementação: F05. Dynamic Markdown Table Engine

**Pré-requisitos:**
- Go 1.22 ou superior instalado no ambiente de desenvolvimento
- Dependências Go declaradas em `go.mod`:
  - `github.com/mattn/go-runewidth` (v0.0.15+)
  - `github.com/charmbracelet/lipgloss` (v0.9+)
- Módulos `internal/buffer` (F01) e `internal/theme` (F02) implementados ou estruturados

---

### Fase 1: Modelos de Dados de Tabela e Detecção de Blocos

**1. Estruturas de Células, Linhas e Tabelas** - Criar em `internal/table/table.go` os tipos `Cell`, `Row`, `Table` e as constantes de enumeração para alinhamento à esquerda, centro e direita.

**2. Parser e Detector de Blocos de Tabela** - Implementar em `internal/table/parser.go` os algoritmos para identificar limites de tabelas no texto e extrair cabeçalhos, delimitadores e linhas de dados.

---

### Fase 2: Cálculo de Larguras Unicode e Formatação GFM

**3. Cálculo de Larguras com Suporte a Emojis** - Desenvolver em `internal/table/formatter.go` a computação da largura visual máxima de cada coluna utilizando `mattn/go-runewidth` para suportar caracteres acentuados e emojis.

**4. Motor de Formatação e Padding GFM** - Implementar a geração de linhas em Markdown padrão com espaços de preenchimento uniformes preservando os pipes e delimitadores.

**5. Aplicação das Diretivas de Alinhamento** - Ajustar a distribuição de padding com base nas diretivas `:---`, `:---:` e `---:` definidas na linha separadora.

---

### Fase 3: Renderizador Visual de Bordas Estéticas Unicode

**6. Renderização de Bordas de Caixa Unicode** - Construir em `internal/table/renderer.go` a geração de molduras estéticas com caracteres de desenho de caixa para topo, divisórias e rodapé da tabela.

**7. Alternância e Fallback para Modo ASCII** - Implementar a opção de renderização com pipes simples e traços quando o modo ASCII estiver ativo nas configurações do editor.

**8. Aplicação de Estilos do Tema Lipgloss** - Integrar as cores e destaques do cabeçalho e bordas utilizando o `CompiledTheme`.

---

### Fase 4: Navegação Interativa e Mutação de Colunas

**9. Navegação de Células com Tab e Shift+Tab** - Implementar em `internal/table/navigation.go` a movimentação inteligente de cursor entre células sem necessidade de mover coluna por coluna manualmente.

**10. Criação Automática de Novas Linhas** - Desenvolver a inserção instantânea de nova linha vazia de tabela ao pressionar `Tab` na última célula da última linha.

**11. Inserção e Deleção Rápida de Colunas** - Criar as operações de atalho para adicionar e remover colunas em todas as linhas da tabela sincronizadamente.

**12. Integração com o Modo Insert** - Conectar os comandos de navegação e mutação de tabela ao despachante de eventos de teclado do editor no modo de inserção.

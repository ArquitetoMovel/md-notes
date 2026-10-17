# Plano de Implementação: F07. TUI Viewport & Status Line Renderer

**Pré-requisitos:**
- Go 1.22 ou superior instalado no ambiente de desenvolvimento
- Dependências Go declaradas em `go.mod`:
  - `github.com/charmbracelet/bubbletea` (v0.25+)
  - `github.com/charmbracelet/lipgloss` (v0.9+)
- Módulos `internal/buffer` (F01), `internal/theme` (F02) e `internal/vim` (F03) implementados ou estruturados

---

### Fase 1: Componentes Laterais e Controladores Visuais

**1. Renderizador de Números de Linha** - Criar em `internal/ui/gutter.go` a renderização da coluna lateral com suporte a modos de numeração absoluta e relativa com largura adaptativa.

**2. Barra de Rolagem Minimalista** - Implementar em `internal/ui/scrollbar.go` o indicador visual de posição de rolagem na margem direita da janela.

**3. Controlador de Formato de Cursor ANSI** - Desenvolver em `internal/ui/cursor.go` a emissão de sequências de escape para alternar os cursores de bloco, barra e sublinhado no terminal.

---

### Fase 2: Barra de Status Inferior Elegante com Lipgloss

**4. Estrutura de Segmentos da Barra de Status** - Criar em `internal/ui/statusline.go` a composição dos blocos temáticos com Lipgloss e distribuição espacial uniforme.

**5. Indicadores Dinâmicos de Modo e Arquivo** - Implementar a exibição em tempo real do modo ativo com cores do tema, caminho do arquivo, indicador de modificação `[+]`, coordenadas do cursor e percentual de navegação.

**6. Integração com Indicadores de Busca e Erros** - Adicionar à barra de status a contagem de ocorrências de busca e o espaço reservado para mensagens informativas e erros do editor.

---

### Fase 3: Motor de Viewport e Soft-Wrap Inteligente

**7. Cálculo de Dimensões e Linhas Visíveis** - Desenvolver em `internal/ui/viewport.go` o particionamento da janela do terminal e a extração exclusiva das linhas contidas na área de visualização.

**8. Algoritmo de Soft-Wrap e Coordenadas** - Implementar a quebra visual suave de linhas extensas e a matriz de mapeamento bidirecional entre coordenadas do buffer e da tela.

**9. Controle de Rolagem com Scrolloff** - Desenvolver o ajuste automático de rolagem vertical mantendo a margem de visibilidade ao redor do cursor.

---

### Fase 4: Integração com Bubble Tea e Otimização a 60 FPS

**10. Tratamento de Redimensionamento do Terminal** - Conectar a mensagem `tea.WindowSizeMsg` no método `Update` do modelo para recalcular dinamicamente as dimensões de tela.

**11. Composição Final de Tela** - Integrar no método `View` a junção harmoniosa do gutter, viewport, scrollbar e barra de status.

**12. Otimização de Renderização a 60 FPS** - Garantir que o pipeline de desenho completo execute em tempo inferior a 16 milissegundos por quadro.

# Plano de Implementação: F08. Interactive Configuration Panel & Settings UI

**Pré-requisitos:**
- Go 1.22 ou superior instalado no ambiente de desenvolvimento
- Dependências Go declaradas em `go.mod`:
  - `github.com/charmbracelet/bubbletea` (v0.25+)
  - `github.com/charmbracelet/lipgloss` (v0.9+)
  - `github.com/pelletier/go-toml/v2` (v2.2+)
- Módulos `internal/config` (F02), `internal/theme` (F02), `internal/vim` (F03) e `internal/ui` (F07) estruturados e funcionais

---

### Fase 1: Motor de Persistência Atômica e Comandos Ex

**1. Persistência Atômica de Configuração** - Desenvolver a gravação segura do arquivo de configuração serializado em TOML, utilizando criação de arquivo temporário com sincronização em disco e renomeação atômica. Consultar a especificação para convenções de diretório e nomenclatura.

**2. Comandos Ex de Configuração** - Estender o despachante de comandos do motor modal para reconhecer os comandos de abertura do painel de configurações e expor o callback de acionamento. Consultar a especificação para os nomes dos comandos e resultados esperados.

**3. Coordenação com Monitor de Arquivo** - Integrar a pausa temporária do monitor de eventos do arquivo de configuração durante a escrita interna para prevenir recarregamentos desnecessários e loops concorrentes. Consultar a especificação para os métodos de coordenação.

---

### Fase 2: Componente de Interface Gráfica e Navegação

**4. Estrutura de Estado do Painel** - Criar no componente de interface a gestão de dados necessária para controlar abas ativas, lista de temas disponíveis, opções de preferências do editor e tokens de cores. Consultar a especificação para as opções e campos abrangidos.

**5. Navegação por Teclado e Alternância de Abas** - Implementar o tratamento de eventos de entrada para transição cíclica entre abas e navegação vertical suave entre as opções configuráveis. Consultar a especificação para os atalhos de teclado mapeados.

**6. Controles de Opções e Edição de Cores** - Desenvolver a alternância de valores de preferências do editor e o fluxo de edição de valores hexadecimais com validação sintática. Consultar a especificação para as regras de validação e ciclos de valores.

---

### Fase 3: Renderização Visual e Composição de Tela

**7. Renderização da Caixa de Diálogo com Lipgloss** - Construir o desenho visual da janela modal com moldura estilizada pelo tema ativo, cabeçalho de abas, área de conteúdo e rodapé com atalhos. Consultar a especificação para regras de dimensionamento e proporções.

**8. Algoritmo de Sobreposição Linha a Linha** - Implementar a mesclagem visual da caixa modal centralizada sobre as linhas de texto do documento de fundo, preservando o conteúdo visível ao redor. Consultar a especificação para o algoritmo de cálculo de coordenadas e fatiamento.

**9. Tratamento Visual de Mensagens e Alertas** - Integrar à caixa de diálogo e à barra de status os indicadores informativos de erro de gravação em disco e avisos de dimensões insuficientes de terminal. Consultar a especificação para os textos e locais de exibição.

---

### Fase 4: Orquestração no Bubble Tea e Live Preview

**10. Invocação e Atalho Global no Modelo** - Conectar no modelo principal da aplicação o acionamento do painel a partir dos comandos modais e do atalho de função global, assegurando a verificação prévia de dimensões da janela. Consultar a especificação para as mensagens e regras de bloqueio.

**11. Mecanismo de Live Preview e Snapshot** - Implementar a captura do estado de configuração original na abertura, a atualização visual dinâmica do documento durante a navegação e a restauração fiel em cancelamentos. Consultar a especificação para o fluxo de backup e reversão.

**12. Ciclo de Salvamento e Composição Final** - Orquestrar o processo de persistência com confirmação na barra de status e incorporar a sobreposição da janela modal no método central de renderização da interface. Consultar a especificação para a sequência de finalização do ciclo.

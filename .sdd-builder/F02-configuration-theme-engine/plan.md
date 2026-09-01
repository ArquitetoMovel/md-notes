# Plano de Implementação: F02. Configuration & Theme Engine

**Pré-requisitos:**
- Go 1.22 ou superior instalado no ambiente de desenvolvimento
- Dependências Go declaradas em `go.mod`:
  - `github.com/charmbracelet/lipgloss` (v0.9+)
  - `github.com/pelletier/go-toml/v2` (v2.2+)
  - `github.com/fsnotify/fsnotify` (v1.7+)

---

### Fase 1: Resolução de Configuração XDG e Parser TOML

**1. Estruturas de Configuração e Resolução de Caminhos XDG** - Definir em `internal/config/config.go` os modelos de dados para preferências do editor e resolução dos caminhos do arquivo de configuração nos padrões XDG e AppData por sistema operacional.

**2. Geração Automática do Arquivo Padrão** - Implementar em `internal/config/loader.go` a criação automática de `config.toml` com documentação explicativa detalhada caso o arquivo não seja encontrado na inicialização.

**3. Parser TOML e Tratamento Resiliente de Falhas** - Desenvolver a lógica de deserialização segura do arquivo TOML, assegurando fallback automático e aviso não bloqueante para o tema padrão em caso de erros de sintaxe.

---

### Fase 2: Paletas de Cores Embutidas e Compilação de Estilos Lipgloss

**4. Paletas dos Temas Pré-definidos** - Registrar em `internal/theme/palette.go` as 7 paletas de cores incorporadas (`dracula`, `nord`, `catppuccin-mocha`, `catppuccin-macchiato`, `monokai`, `default-dark` e `default-light`).

**5. Sobreposição Granular de Cores Customizadas** - Implementar no módulo de temas a mesclagem de cores customizadas em formato Hex (`#RRGGBB`) ou ANSI 256 sobre a paleta selecionada.

**6. Compilador de Estilos de Alta Eficiência** - Construir em `internal/theme/compiler.go` o compilador de estilos Lipgloss para gerar a struct imutável de estilos pré-compilados com tempo de execução inferior a 5 milissegundos.

---

### Fase 3: Detecção de Terminal e Hot-Reload Dinâmico

**7. Detecção Automática de Modo Claro/Escuro** - Desenvolver em `internal/theme/detect.go` a leitura de variáveis de ambiente do terminal para resolver a opção de tema automático baseada no fundo da janela.

**8. Serviço de Hot-Reload de Configuração** - Criar em `internal/config/watcher.go` a rotina de monitoramento do arquivo `config.toml` via eventos de sistema para detectar alterações em disco com controle de debounce.

**9. Despacho de Eventos para o Ciclo de Renderização** - Integrar o disparo de mensagens de recarregamento de tema no modelo da aplicação para atualizar instantaneamente os componentes visuais em execução.

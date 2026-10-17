# Plano de Implementação: F03. Modal Editing Engine (Vim Emulation)

**Pré-requisitos:**
- Go 1.22 ou superior instalado no ambiente de desenvolvimento
- Dependências Go declaradas em `go.mod`:
  - `github.com/charmbracelet/bubbletea` (v0.25+)
  - `github.com/atotto/clipboard` (v0.1.4+)
- Módulo `internal/buffer` (F01) implementado e testado

---

### Fase 1: Tipos de Estado Modal, Clipboard e Pilha de Histórico

**1. Modelos de Estado Modal e Enum de Modos** - Criar em `internal/vim/state.go` as definições dos modos de operação, estruturas de posição de cursor, seleção visual e registradores de cópia e corte.

**2. Integração com a Área de Transferência** - Implementar em `internal/vim/clipboard.go` o serviço de leitura e escrita com o clipboard do sistema operacional mantendo fallback em memória para ambientes headless.

**3. Pilha de Histórico de Undo e Redo** - Construir em `internal/vim/history.go` a estrutura de histórico circular com até 200 snapshots imutáveis do buffer para operações de reversão determinística.

---

### Fase 2: Motor de Movimento e Navegação UTF-8

**4. Navegação Direcional e Limites de Documento** - Desenvolver em `internal/vim/motion.go` as funções de movimentação direcional do cursor sobre coordenadas de runes UTF-8 e saltos para início e fim de linha e documento.

**5. Navegação por Palavras e Símbolos** - Implementar os algoritmos de avanço e recuo de palavra considerando delimitações de caracteres alfanuméricos e pontuações do Markdown.

**6. Busca Inline de Caracteres na Linha** - Criar os mecanismos de busca na linha ativa para saltar até ou antes de caracteres específicos à frente ou atrás do cursor.

**7. Multiplicadores Numéricos de Repetição** - Desenvolver o acumulador de dígitos para permitir a repetição dinâmica de movimentos e operações com contadores prefixados.

---

### Fase 3: Operadores de Edição, Modos Visuais e Linha de Comandos Ex

**8. Transições de Entrada no Modo Insert** - Implementar no motor modal as transições para o modo de inserção nas posições adequadas e o registro de checkpoints de histórico ao retornar para o modo normal.

**9. Operadores de Deleção, Cópia e Colagem** - Desenvolver os comandos de exclusão de caracteres, linhas e palavras com sincronização no registrador e colar após ou antes do cursor.

**10. Seleção nos Modos Visuais** - Implementar a seleção de texto por caractere, por linha e em bloco retangular com suporte a corte e cópia para o clipboard.

**11. Interpretador da Linha de Comandos Ex** - Construir em `internal/vim/command.go` o analisador sintático de comandos iniciados com dois pontos para salvar, sair e navegar entre linhas.

---

### Fase 4: Integração com o Modelo Bubble Tea e Tratamento de Erros

**12. Integração da Máquina de Estados no Ciclo Bubble Tea** - Integrar o despachante de eventos do Vim no método de atualização do modelo da aplicação para gerenciar a captura de teclado.

**13. Linha de Comando Inferior e Indicadores** - Ajustar a visualização da aplicação para exibir o prompt de digitação de comandos e alternar os indicadores dos modos ativos.

**14. Formatação de Mensagens de Erro Padrão Vim** - Implementar as notificações visuais para comandos não reconhecidos, limites de histórico de alterações e avisos de saída com alterações não salvas.

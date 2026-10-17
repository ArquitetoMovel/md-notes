# Plano de Implementação: F09. Seleção com mouse e transferência de texto

**Pré-requisitos:**
- F01 (buffer), F03 (edição modal e histórico) e F07 (viewport) disponíveis no código.
- Go e dependências do projeto configurados; atualização autorizada do Bubble Tea para uma versão v0/v1 com suporte a colagem delimitada.
- PRD e contratos técnicos desta feature em `docs/PRD.md` e `spec.md`.

### Fase 1: Entrada do terminal e coordenadas visuais

**1. Compatibilidade do Bubble Tea** - Atualizar a dependência dentro da linha v0/v1 e confirmar que eventos de colagem delimitada chegam ao modelo como uma operação única, mantendo entrada por TTY e pipe. Seguir o contrato de eventos de `spec.md`.

**2. Mapeamento do viewport** - Completar a conversão entre células do terminal e posições lógicas do documento para gutter, rolagem, quebra visual e caracteres de largura variável. Usar o mesmo critério visual no desenho e na interação do mouse.

**3. Ciclo do arraste** - Ligar pressionamento, movimento e liberação do botão primário à seleção visual do motor Vim, preservando a rolagem já existente e a prioridade do painel de configurações.

### Fase 2: Cópia e integração com a área de transferência

**4. Extração do trecho selecionado** - Garantir que a seleção produza o texto bruto do buffer, com marcadores Markdown e quebras lógicas, sem inserir quebras artificiais de `word_wrap`.

**5. Operação de clipboard** - Oferecer leitura e escrita que informem falhas à interface sem alterar o comportamento dos comandos Vim existentes que usam o registrador interno.

**6. Atalho de cópia** - Priorizar a cópia quando houver seleção, apresentar confirmação ou erro na barra de status e conservar o comportamento de saída protegida de `Ctrl+C` quando não houver seleção.

### Fase 3: Colagem e histórico

**7. Substituição no buffer** - Criar uma operação única de inserção ou substituição que preserve o texto não selecionado, as terminações de linha e o estado do documento conforme `spec.md`.

**8. Entrada literal de paste** - Encaminhar eventos de colagem delimitada e atalhos de colar para a operação de substituição, sem interpretar caracteres colados como comandos de edição.

**9. Limites e falhas** - Validar conteúdo e tamanho antes da mutação; manter documento e seleção intactos nos quatro cenários de erro definidos no PRD.

**10. Desfazer e modos** - Fazer da colagem uma etapa independente no histórico, incluindo quando havia digitação anterior no mesmo período de INSERT. Conservar Visual Block em NORMAL sem seleção e bloquear colagem no painel, em COMMAND e em SEARCH.

### Fase 4: Integração e acabamento

**11. Feedback visual** - Ajustar destaque, cursor e barra de status para que o intervalo selecionado permaneça visível após soltar o mouse e desapareça ao cancelar, copiar ou substituir conforme o fluxo do PRD.

**12. Consolidação do fluxo** - Revisar os caminhos de mouse, clipboard, buffer e viewport juntos, incluindo arquivos com UTF-8 e terminações CRLF, e atualizar a documentação de uso dos atalhos condicionados ao terminal.

### Extensão: botão direito do mouse

**13. Cópia contextual** - Tratar pressionamento do botão direito com seleção ativa como cópia estrita, preservando seleção e buffer em caso de falha.

**14. Colagem contextual** - Sem seleção, converter a célula clicada em posição de inserção, inclusive na área vazia abaixo do documento, e colar em NORMAL ou INSERT usando a validação e o checkpoint único existentes.

**15. Cobertura e documentação** - Verificar erros, modos bloqueados, barra de status e undo; documentar o gesto no README.

# Painel de configurações

## Contexto Atual

Atualmente já existe suporte a configuração através do arquivo *TOML*.
Porém é necessário o usuário sair do fluxo, editar e configurar o arquivo. 

## Proposta 

A proposta é criar uma feature com opção de comando dentro do **mdn** `:c` e ou `:config`, que quando acionado irá exibir um _popup_ centralizado no editor permitindo selecionar entre temas fontes etc sem ter que manipular o arquivo de configurações diretamente.
O arquivo *TOML* continuará existindo e as configurações deverão ser obtidas por meio do arquivo.
As configurações internas servirá como uma interface mais amigavél para editar o arquivo.

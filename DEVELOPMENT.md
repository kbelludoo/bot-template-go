# Validação do template

Revisão local de infraestrutura em 20/10/2026. Os testes usam servidores HTTP e
WebSocket locais; não precisam de credenciais nem iniciam contas no jogo.

```sh
go test ./...
go vet ./...
go test -race ./...
go run ./cmd/build linux
```

O detector `-race` exige CGO e um compilador C compatível. Se não estiverem
disponíveis, execute essa etapa em um ambiente configurado antes de distribuir
uma versão. Testes repetidos não substituem o detector de condições de corrida.

## Configuração e sessões

- O exemplo começa com `enabled: false` e `autoHunt: false`. Configure a conta e
  audite o protocolo do jogo antes de ativá-la.
- A seleção por `ACCOUNT` continua permitindo selecionar uma conta desativada.
  `autoHunt` permanece independente dessa seleção.
- IDs vazios ou duplicados, logins/tokens repetidos e `bindIP` inválido impedem a
  inicialização. Tokens diferentes da mesma conta ainda exigem identificação no
  servidor; essa validação local não consegue provar que são contas distintas.
- Com várias contas, `/api/ticket` exige `accountId`. IDs desconhecidos geram erro,
  sem entregar a credencial à primeira conta.
- `accounts.json` novo usa permissões `0600` em sistemas que as suportam. Erros de
  leitura não sobrescrevem a configuração existente.

## Proteção e recuperação

HTTP 403/429 ou rejeições reconhecidas desarmam o circuit breaker. O callback
desliga `AutoHunt` e marca `cooldown` antes de retornar. As requisições HTTP são
serializadas por conta para que ações enfileiradas não ultrapassem o bloqueio.

Há uma única tentativa de `POST hunt/stop`, sem progresso e sem retry, após a
primeira rejeição. Isso usa a operação de saída já existente no adaptador; sua
eficácia no servidor precisa ser confirmada em uma homologação do protocolo.
Nenhuma mudança arbitrária de localização é enviada como save.

Após o cooldown, apenas uma leitura `GET state` com JSON válido e nível positivo
pode concluir a recuperação. Saves, combate e demais ações ficam bloqueados em
HALF-OPEN. Falhas na leitura reabrem o circuito com backoff. Respostas antigas não
cancelam a proteção. Após recuperar o estado, a caça exige reativação manual.

Os callbacks são síncronos e executados fora do mutex do circuit breaker. Eles
podem consultar seu estado, mas não devem iniciar requisições pelo mesmo cliente
HTTP enquanto uma requisição desse cliente está em andamento.

`PokeTrueState.RawEstado` mantém campos desconhecidos e números sem perda de
precisão. Se um adaptador implementar saves, deve construir o envelope
`{"estado": raw}` a partir desse mapa, alterando somente os campos necessários.

## Auditoria de stamina/fadiga

1. **Existe limitador de caça?** Os modelos locais e os arquivos públicos
   `game.js` e `nuvem.js` inspecionados não confirmaram um contador de stamina de
   caça. HP e cargas de treino não comprovam esse mecanismo. A ausência de um
   campo nesses arquivos não comprova ausência no servidor. É necessário
   confirmar em estado canônico/pacotes antes de homologar caça contínua.
2. **Quando parar?** O controlador genérico bloqueia operação antes da primeira
   métrica, com stamina em zero ou até 14h, fadiga a partir de 80%, ou farm abaixo
   de dois minutos, inclusive zero. A retomada exige descanso mínimo e, para
   fadiga, valor abaixo de 30% **e** farm acima de dez minutos.
3. **Como recuperar?** O controlador administra a histerese e exporta telemetria.
   Saída da caça, zona segura, treino ou recarga precisam ser ligados a operações
   verificadas de cada jogo. O adaptador PokeTrue ainda não integra métricas de
   fadiga nem comprova regeneração; essa revisão não homologa esse fluxo.

## Catálogos e escopo

Os arquivos ausentes foram recuperados das URLs públicas referenciadas pelo
cliente do jogo: [hunts](https://poketrue.net/assets/hunts.json),
[pokedex](https://poketrue.net/assets/pokedex.json) e
[loot](https://poketrue.net/assets/loot.json). O catálogo de evoluções existente
foi preservado. O `.gitignore` permite somente os JSONs de `pkg/game/data`;
credenciais e dados locais continuam ignorados.

Os módulos de quests, equipamentos, treino, inventário, grupo, runas, evolução,
ginásios e chefes já existentes continuam precisando de homologação contra o
protocolo atual. Os testes desta revisão validam compilação, catálogo e
infraestrutura. Não representam auditoria exaustiva de gameplay ou garantia de
compatibilidade com todos os subsistemas do servidor.

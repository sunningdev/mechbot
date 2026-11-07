import type { Post } from "./posts";

export const ptPosts: Post[] = [
  {
    slug: "health-factor",
    title: "Como o fator de saúde é calculado? Entendendo o número “até a liquidação”",
    summary:
      "Aave, Compound e Morpho dão nomes diferentes ao limite de liquidação, mas a ideia é a mesma. Uma fórmula explica tudo e mostra quanto os preços ainda podem cair.",
    date: "2026-10-28",
    tag: "Fundamentos",
    readMin: 5,
    body: [
      {
        type: "p",
        text: "Um protocolo de empréstimo não se importa com quanto você pegou emprestado. Ele se importa com uma coisa só: depois de descontar seu colateral pelo limite de liquidação, ele ainda cobre sua dívida? O fator de saúde (Health Factor, HF) é exatamente essa razão.",
      },
      { type: "h", text: "Uma fórmula" },
      { type: "code", text: "HF = Σ(valor do colateral × limite de liquidação) / valor da dívida" },
      {
        type: "p",
        text: "Acima de 1 é seguro. Em 1 você chega ao limite de liquidação. Abaixo de 1, qualquer pessoa pode liquidar sua posição e ficar com parte do seu colateral como bônus.",
      },
      { type: "h", text: "Três protocolos, três nomes" },
      {
        type: "list",
        items: [
          "Aave V3: cada ativo tem um Liquidation Threshold, e o contrato retorna o fator de saúde diretamente.",
          "Compound V3: cada ativo de colateral tem um liquidateCollateralFactor; nós mesmos aplicamos a mesma fórmula.",
          "Morpho Blue: cada mercado tem um LLTV (empréstimo sobre valor na liquidação). Mesma fórmula, com um único ativo de colateral.",
        ],
      },
      { type: "h", text: "O que significa “até a liquidação”" },
      {
        type: "p",
        text: "No painel, “até a liquidação” é quanto o preço do seu colateral pode cair, com a dívida inalterada, antes de a liquidação ser acionada:",
      },
      { type: "code", text: "Até a liquidação = 1 − 1 / HF" },
      {
        type: "p",
        text: "Com HF 1.25, seu colateral ainda pode cair 20%. Com HF 1.07, sobra só cerca de 6.5% — basta um candle vermelho grande.",
      },
      {
        type: "note",
        text: "Se você pegou emprestado um ativo volátil (por exemplo, ETH), sua dívida também oscila, então a margem real será diferente. Use esse número como referência intuitiva, não como previsão exata.",
      },
      { type: "h", text: "Onde colocar seu limite de alerta" },
      {
        type: "p",
        text: "Com stablecoins de colateral e dívida em stablecoins, dá para ficar mais baixo, por volta de 1.1. Com ETH ou BTC de colateral e dívida em stablecoins, 1.3 a 1.5 é uma boa faixa que deixa tempo para reforçar a garantia.",
      },
    ],
  },
  {
    slug: "rpc-failover",
    title: "O que acontece quando um nó cai? O failover multinó do MechBot",
    summary:
      "RPCs públicos limitam requisições, dão timeout e caem. Veja como os alertas continuam chegando quando os nós falham — e por que “não consegui ler” nunca é tratado como “seguro”.",
    date: "2026-10-26",
    tag: "Engenharia",
    readMin: 6,
    body: [
      {
        type: "p",
        text: "A pior falha de um alerta de liquidação não é um erro, é o silêncio: um nó cai, nada é lido, nada é enviado, e o usuário acha que está tudo bem. O failover do MechBot é construído em torno de uma regra: dado que não conseguimos ler nunca é considerado seguro.",
      },
      { type: "h", text: "Cada rede tem um grupo de nós" },
      {
        type: "p",
        text: "Cada rede tem vários endpoints RPC: primeiro os que você configurou, depois reservas públicas embutidas. Cada consulta tenta em ordem, lembra qual funcionou e começa por ele na próxima vez.",
      },
      { type: "code", text: "Nó configurado → reserva 1 → reserva 2 → …\nSucesso: lembrar este nó\nFalha: passar para o próximo" },
      { type: "h", text: "Quando todos os nós falham" },
      {
        type: "list",
        items: [
          "A posição é marcada como “consulta falhou, nova tentativa na próxima rodada” no painel — nunca aparece como segura.",
          "Se uma fonte de dados falhar para todos os endereços por várias rodadas seguidas, um alerta de plantão vai para o Telegram do operador.",
          "Quando o nó se recupera, um aviso de recuperação confirma que o caminho voltou a funcionar.",
        ],
      },
      { type: "h", text: "O programa também precisa aguentar" },
      {
        type: "p",
        text: "Leitura de protocolos, polling, varredura de pagamentos e tratamento de mensagens do bot têm, cada um, sua própria proteção contra travamentos. Dados ruins de uma posição não derrubam uma rodada inteira, muito menos o serviço.",
      },
      {
        type: "note",
        text: "Quando a API do Telegram retorna um erro, o token do bot é removido do log para que segredos nunca vazem para os sistemas de logging.",
      },
    ],
  },
  {
    slug: "no-spam",
    title: "Por que o MechBot não faz spam: a vida de um alerta",
    summary:
      "Centenas de notificações por dia é o mesmo que nenhuma. Só avisamos quando você cruza seu limite de alerta, e nunca duas vezes em uma hora para a mesma posição, a menos que ela esteja prestes a ser liquidada.",
    date: "2026-10-24",
    tag: "Produto",
    readMin: 4,
    body: [
      {
        type: "p",
        text: "Muitos bots de monitoramento disparam a cada oscilação de preço. Os usuários silenciam e depois perdem justamente o alerta que importa. O MechBot busca o oposto: quieto por padrão, barulhento só antes da liquidação.",
      },
      { type: "h", text: "Verificar: a cada 5 minutos" },
      {
        type: "p",
        text: "A cada 5 minutos o backend lê as posições de todos os seus endereços em cada protocolo, direto dos contratos on-chain, sem APIs de dados de terceiros.",
      },
      { type: "h", text: "Decidir: um único limite" },
      {
        type: "list",
        items: [
          "Sem dívida, ou fator de saúde acima do seu limite: nada é enviado.",
          "Cai abaixo do limite: mensagem imediata com protocolo, rede, fator de saúde, colateral e dívida.",
          "Continua abaixo: sem repetir por 1 hora; se depois de uma hora ainda não se recuperou, mais um lembrete.",
          "Fator de saúde abaixo de 1.05, à beira da liquidação: lembretes a cada 15 minutos até você agir ou a posição se recuperar.",
          "Volta acima do limite: uma única mensagem de “seguro novamente”, e silêncio outra vez.",
        ],
      },
      { type: "code", text: "Seguro ──cai──▶ Alerta (mensagem)\n  ▲             │ lembrete a cada hora (15 min abaixo de 1.05)\n  └──recupera── Seguro novamente (uma mensagem)" },
      { type: "h", text: "O resultado" },
      {
        type: "p",
        text: "Uma posição que vai de segura a perigosa e volta normalmente custa só duas mensagens. Cada uma vale a pena parar o que você está fazendo para olhar.",
      },
    ],
  },
];

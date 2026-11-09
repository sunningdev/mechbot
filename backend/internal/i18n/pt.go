package i18n

var pt = Text{
	DangerHead: `⚠️ MechBot: alerta de liquidação

Protocolo: %s
Endereço: %s
Health factor: %.4f (limite %.2f)
Colateral: %s
Dívida: %s
`,
	ETHDrop:    "Liquidado se o ETH cair mais %.1f%% (preço de liquidação do ETH %s)\n",
	DangerFoot: "\nUma posição pode ser liquidada quando o health factor fica abaixo de 1.0.",

	EarlyHead: `⏱ MechBot: alerta antecipado (Pro)

Protocolo: %s
Endereço: %s
Health factor on-chain: %.4f
Estimado com preços da Binance em tempo real: %.4f (limite %.2f)
`,
	EarlyETH:  "ETH: %s → %s (%+.1f%%)\n",
	EarlyNote: "\nOs oráculos on-chain costumam ficar atrás das exchanges; quando atualizarem, você vai cruzar o seu limite. Ainda dá tempo de agir.\n",

	RateHigh: `📈 MechBot: alerta de taxa de empréstimo (Pro)

Protocolo: %s
Endereço: %s
APY de empréstimo de %s: %.2f%%
Dívida: %s, cerca de %s a mais por ano nessa taxa

Uma alta brusca da taxa costuma indicar que o pool está quase todo emprestado. Seu loop pode já estar no prejuízo; considere pagar parte da dívida ou mudar de mercado.`,
	BorrowWord: "da dívida",
	RateCalm: `✅ MechBot: a taxa de empréstimo caiu

Protocolo: %s
Endereço: %s
APY de empréstimo: %.2f%%`,

	DepegSideCollateral:   "Colateral",
	DepegSideDebt:         "Dívida",
	DepegEffectCollateral: "Seu colateral está valendo menos",
	DepegEffectDebt:       "Sua dívida está crescendo",
	DepegEffectFavorable:  "Por enquanto isso te favorece, mas um depeg costuma vir com oscilações fortes",
	DepegEffectHF:         ", e seu health factor vai cair quando os oráculos atualizarem",
	Depeg: `⚠️ MechBot: alerta de depeg (Pro)

Protocolo: %s
Endereço: %s
%s %s, atrelado a %s, está em %.4f %s (%+.2f%% fora da paridade)
Health factor: %.4f

%s. Considere reduzir a alavancagem: pague parte do empréstimo ou adicione outro colateral.%s`,
	DepegCalm: `✅ MechBot: %s voltou para perto da paridade

Protocolo: %s
Endereço: %s
Agora %.4f %s`,

	AppLine:     "Resolver: %s",
	RescueHead:  "\nO que fazer (voltar a %.2f):\n",
	RescueRepay: "· Pagar %s\n",
	RescueAdd:   "· Ou adicionar %s de colateral\n",
	TokenApprox: "%s %s (cerca de %s)",

	SafeNoDebt: `✅ MechBot: posição segura

Protocolo: %s
Endereço: %s

Este empréstimo foi quitado; não há mais dívida.`,
	Safe: `✅ MechBot: posição segura

Protocolo: %s
Endereço: %s
Health factor: %.4f (limite %.2f)
Colateral: %s
Dívida: %s

De volta acima do seu limite.`,

	ListHead:   "Endereço %s\nLimite %.2f\n",
	ListFailed: "\n%s\nFalha na consulta: %v\n",
	ListPos:    "\n%s\nHealth factor %.4f\nColateral %s\nDívida %s\n",
	ListDrop:   "Liquidado se o ETH cair %.1f%% (%s)\n",
	ListAPY:    "APY de empréstimo %.2f%%\n",
	ListQuiet:  "\nSem dívida: %s\n",
	ListSep:    ", ",

	AddedDebt:   "Adicionado. Posições atuais:\n\n",
	AddedFailed: "Endereço %s adicionado.\n\nAlguns protocolos não responderam agora. Tente /list daqui a pouco.",
	AddedQuiet: `Endereço %s adicionado.

Este endereço não tem empréstimos no momento: verificamos Aave V3, Compound V3 e Morpho Blue (Ethereum · Arbitrum · Optimism), então ele não pode ser liquidado.

Vamos continuar monitorando. Se ele pegar empréstimo nesses protocolos depois, os alertas começam sozinhos; não precisa adicionar de novo.`,
	AddedStarter: "\n\nO Starter só envia alertas da Aave V3. Empréstimos na Compound e na Morpho exigem Pro.",

	StarterLimits: `· Ficam só os alertas da Aave V3; os da Compound e da Morpho param
· Alertas antecipados, de taxa e de depeg param
· O mapa de liquidações passa a ter 1 hora de atraso`,
	ProEndingPaid:  "Seu Pro termina em menos de 24 horas",
	ProEndingTrial: "Seu teste de 7 dias do Pro termina em menos de 24 horas",
	ProEnding: `⏳ MechBot: %s

Depois disso você passa para o Starter (grátis):
· Só os seus %d primeiros endereços seguem monitorados; o resto fica pausado (não é apagado)
%s

Manter o Pro: %s`,
	ProEndedPaid:   "O Pro terminou; agora você está no Starter",
	ProEndedTrial:  "Seu teste do Pro terminou; agora você está no Starter",
	ProEndedWatch:  "· Seus %d primeiros endereços continuam monitorados",
	ProEndedPaused: "; os outros %d estão pausados (dados mantidos, voltam ao renovar)",
	ProEnded: `MechBot: %s

%s
%s

Renove para recuperar tudo na hora: %s`,

	StatusStarter: "Starter (grátis para sempre, só alertas da Aave V3)",
	StatusPro:     "Pro (todos os protocolos)",
	StatusTrial:   "Teste do Pro (todos os protocolos, volta ao Starter no fim)",
	StatusNone:    "Nenhum",
	Status: `Conta MechBot

Plano: %s
Endereços: %d/%d
Pro até: %s
Convidado por: %s`,

	PlanMonth:   "Pro mensal",
	PlanQuarter: "Pro trimestral",
	PlanYear:    "Pro anual",
	PayTooLow:   "⚠️ Recebemos %s USDT (%s)\n\nÉ menos que o plano mais barato (%d USDT), então o Pro não foi ativado. Se tiver dúvidas, responda ao bot.",
	PayBonus:    "\n🎁 Bônus de convite +%d dias",
	PayDone:     "✅ MechBot Pro ativado\n\nRecebemos %s USDT (%s): %s, %d dias%s\nPro até: %s",
	PayReferrer: "🎁 Recompensa de convite +%d dias\n\nAlguém que você convidou acabou de comprar %s pela primeira vez.\nPro até: %s",

	Help: `MechBot: alertas de liquidação

/add 0x... Monitorar um endereço (Starter 3, Pro 20)
/list Health factor em cada protocolo
/remove 0x... Parar de monitorar um endereço
/status Status da conta
/lang Mudar o idioma das mensagens

Cada endereço começa com limite de 1.2. Você pode mudar no cartão do endereço no console web.

O Starter (grátis para sempre) envia alertas da Aave V3 (Ethereum, Arbitrum, Optimism). O Pro adiciona Compound V3 e Morpho Blue.`,
	HelpConsole:       "\n\nConsole web: %s",
	TrialNote:         "\n\n🎁 Novos usuários ganham %d dias de Pro: alertas em todos os protocolos, alertas antecipados e o mapa de liquidações em tempo real. Nada é cobrado no fim; você só volta para o Starter gratuito.",
	UnknownCommand:    "Esse comando não existe.",
	RegisterFailed:    "Não foi possível concluir o cadastro. Tente mais tarde.",
	Registered:        "Cadastro concluído.",
	AlreadyRegistered: "Você já está cadastrado.",
	QueryFailed:       "A consulta falhou. Tente mais tarde.",
	AddUsage:          "Uso: /add seguido de um endereço de carteira que começa com 0x",
	AddFailed:         "Não foi possível adicionar. Tente mais tarde.",
	AddExists:         "Este endereço já está sendo monitorado.",
	AddLimit:          "Seu plano permite até %d endereços.",
	RemoveUsage:       "Uso: /remove seguido de um endereço de carteira que começa com 0x",
	Removed:           "Removido %s",
	RemoveMissing:     "Esse endereço não está na sua lista.",
	RemoveFailed:      "Não foi possível remover. Tente mais tarde.",
	ThresholdMoved:    "Agora o limite é definido por endereço (padrão 1.2). Mude no cartão do endereço no console web.",
	NoAddresses:       "Nenhum endereço ainda. Adicione um com /add 0x...",
	Paused:            "\n\n(Acima do limite do Starter, alertas pausados)",

	LoginPrivate: "Confirme o login em uma conversa privada com o bot.",
	LoginInvalid: "Este link de login é inválido ou expirou. Volte ao site e toque de novo em “Entrar com Telegram”.",
	LoginPrompt: `Você está entrando no console web do MechBot.

Código mostrado no site: %s

Só toque em “Confirmar login” se os códigos forem iguais. Se você não acabou de tentar entrar no site, não confirme.`,
	LoginButton:       "Confirmar login",
	LoginExpiredToast: "Link expirado",
	LoginExpired:      "Este link de login expirou. Volte ao site e toque de novo em “Entrar com Telegram”.",
	LoginFailedToast:  "Falha no login, tente de novo",
	LoginUsedToast:    "Link não é mais válido",
	LoginUsed:         "Este link de login já foi usado ou expirou.",
	LoginConfirmToast: "Confirmado",
	LoginConfirmed:    "Login confirmado. Volte ao site; o console abre sozinho.",

	LangPrompt: "Escolha o idioma das mensagens do bot:",
	LangSet:    "Pronto. A partir de agora as mensagens chegam em português.",
}

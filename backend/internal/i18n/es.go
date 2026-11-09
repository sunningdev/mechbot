package i18n

var es = Text{
	DangerHead: `⚠️ MechBot: alerta de liquidación

Protocolo: %s
Dirección: %s
Health factor: %.4f (umbral %.2f)
Colateral: %s
Deuda: %s
`,
	ETHDrop:    "Se liquida si ETH cae otro %.1f%% (precio de liquidación de ETH %s)\n",
	DangerFoot: "\nUna posición puede ser liquidada cuando su health factor baja de 1.0.",

	EarlyHead: `⏱ MechBot: alerta anticipada (Pro)

Protocolo: %s
Dirección: %s
Health factor on-chain: %.4f
Estimado con precios de Binance en tiempo real: %.4f (umbral %.2f)
`,
	EarlyETH:  "ETH: %s → %s (%+.1f%%)\n",
	EarlyNote: "\nLos oráculos on-chain suelen ir por detrás de los exchanges; cuando se actualicen, cruzarás tu umbral. Todavía estás a tiempo.\n",

	RateHigh: `📈 MechBot: alerta de tipo de préstamo (Pro)

Protocolo: %s
Dirección: %s
APY de préstamo de %s: %.2f%%
Deuda: %s, unos %s extra al año a este tipo

Una subida brusca del tipo suele significar que el pool se ha quedado sin liquidez. Tu loop puede estar ya en pérdidas; considera devolver una parte o cambiar de mercado.`,
	BorrowWord: "la deuda",
	RateCalm: `✅ MechBot: el tipo de préstamo ha bajado

Protocolo: %s
Dirección: %s
APY de préstamo: %.2f%%`,

	DepegSideCollateral:   "Colateral",
	DepegSideDebt:         "Deuda",
	DepegEffectCollateral: "Tu colateral vale menos",
	DepegEffectDebt:       "Tu deuda está creciendo",
	DepegEffectFavorable:  "De momento te favorece, pero un desanclaje suele venir con movimientos bruscos",
	DepegEffectHF:         ", y tu health factor bajará cuando se actualicen los oráculos",
	Depeg: `⚠️ MechBot: alerta de desanclaje (Pro)

Protocolo: %s
Dirección: %s
%s %s, que sigue a %s, está en %.4f %s (%+.2f%% fuera del ancla)
Health factor: %.4f

%s. Considera reducir el apalancamiento: devuelve parte del préstamo o añade otro colateral.%s`,
	DepegCalm: `✅ MechBot: %s vuelve a estar cerca de su ancla

Protocolo: %s
Dirección: %s
Ahora %.4f %s`,

	AppLine:     "Resolver: %s",
	RescueHead:  "\nQué hacer (volver a %.2f):\n",
	RescueRepay: "· Devolver %s\n",
	RescueAdd:   "· O añadir %s de colateral\n",
	TokenApprox: "%s %s (unos %s)",

	SafeNoDebt: `✅ MechBot: posición segura

Protocolo: %s
Dirección: %s

Este préstamo está saldado; ya no hay deuda.`,
	Safe: `✅ MechBot: posición segura

Protocolo: %s
Dirección: %s
Health factor: %.4f (umbral %.2f)
Colateral: %s
Deuda: %s

De nuevo por encima de tu umbral.`,

	ListHead:   "Dirección %s\nUmbral %.2f\n",
	ListFailed: "\n%s\nError de consulta: %v\n",
	ListPos:    "\n%s\nHealth factor %.4f\nColateral %s\nDeuda %s\n",
	ListDrop:   "Se liquida si ETH cae %.1f%% (%s)\n",
	ListAPY:    "APY de préstamo %.2f%%\n",
	ListQuiet:  "\nSin deuda: %s\n",
	ListSep:    ", ",

	AddedDebt:   "Añadida. Posiciones actuales:\n\n",
	AddedFailed: "Añadida %s.\n\nAlgunos protocolos no respondieron ahora mismo. Vuelve a probar con /list en un rato.",
	AddedQuiet: `Añadida %s.

Esta dirección no tiene préstamos ahora mismo: revisamos Aave V3, Compound V3 y Morpho Blue (Ethereum · Arbitrum · Optimism), así que no puede ser liquidada.

Seguiremos vigilándola. Si más adelante pide prestado en estos protocolos, las alertas empezarán solas; no hace falta volver a añadirla.`,
	AddedStarter: "\n\nStarter solo avisa en Aave V3. Los préstamos en Compound y Morpho requieren Pro.",

	StarterLimits: `· Solo quedan las alertas de Aave V3; las de Compound y Morpho se detienen
· Se detienen las alertas anticipadas, de tipo de préstamo y de desanclaje
· El mapa de liquidaciones pasa a tener 1 hora de retraso`,
	ProEndingPaid:  "Tu Pro termina en menos de 24 horas",
	ProEndingTrial: "Tu prueba de Pro de 7 días termina en menos de 24 horas",
	ProEnding: `⏳ MechBot: %s

Después pasarás a Starter (gratis):
· Solo se vigilan tus %d primeras direcciones; el resto se pausa (no se borra)
%s

Mantener Pro: %s`,
	ProEndedPaid:   "Pro ha terminado; ahora estás en Starter",
	ProEndedTrial:  "Tu prueba de Pro ha terminado; ahora estás en Starter",
	ProEndedWatch:  "· Tus %d primeras direcciones siguen vigiladas",
	ProEndedPaused: "; las otras %d están en pausa (los datos se conservan y se reanudan al renovar)",
	ProEnded: `MechBot: %s

%s
%s

Renueva y lo recuperas todo al instante: %s`,

	StatusStarter: "Starter (gratis para siempre, solo alertas de Aave V3)",
	StatusPro:     "Pro (todos los protocolos)",
	StatusTrial:   "Prueba de Pro (todos los protocolos, vuelve a Starter al terminar)",
	StatusNone:    "Ninguno",
	Status: `Cuenta de MechBot

Plan: %s
Direcciones: %d/%d
Pro hasta: %s
Invitado por: %s`,

	PlanMonth:   "Pro mensual",
	PlanQuarter: "Pro trimestral",
	PlanYear:    "Pro anual",
	PayTooLow:   "⚠️ Recibidos %s USDT (%s)\n\nEs menos que el plan más barato (%d USDT), así que Pro no se activó. Si tienes dudas, responde al bot.",
	PayBonus:    "\n🎁 Bonus por invitación +%d días",
	PayDone:     "✅ MechBot Pro activado\n\nRecibidos %s USDT (%s): %s, %d días%s\nPro hasta: %s",
	PayReferrer: "🎁 Recompensa por invitación +%d días\n\nAlguien a quien invitaste acaba de comprar %s por primera vez.\nPro hasta: %s",

	Help: `MechBot: alertas de liquidación

/add 0x... Vigilar una dirección (Starter 3, Pro 20)
/list Health factor en cada protocolo
/remove 0x... Dejar de vigilar una dirección
/status Estado de la cuenta
/lang Cambiar el idioma de los mensajes

Cada dirección empieza con un umbral de 1.2. Puedes cambiarlo en la tarjeta de la dirección en la consola web.

Starter (gratis para siempre) avisa en Aave V3 (Ethereum, Arbitrum, Optimism). Pro añade Compound V3 y Morpho Blue.`,
	HelpConsole:       "\n\nConsola web: %s",
	TrialNote:         "\n\n🎁 Los nuevos usuarios reciben %d días de Pro: alertas en todos los protocolos, alertas anticipadas y el mapa de liquidaciones en tiempo real. Al terminar no se cobra nada; simplemente vuelves al plan gratuito Starter.",
	UnknownCommand:    "Ese comando no existe.",
	RegisterFailed:    "No se pudo completar el registro. Inténtalo más tarde.",
	Registered:        "Registro completado.",
	AlreadyRegistered: "Ya estás registrado.",
	QueryFailed:       "La consulta falló. Inténtalo más tarde.",
	AddUsage:          "Uso: /add seguido de una dirección de wallet que empiece por 0x",
	AddFailed:         "No se pudo añadir. Inténtalo más tarde.",
	AddExists:         "Esta dirección ya está vigilada.",
	AddLimit:          "Tu plan permite hasta %d direcciones.",
	RemoveUsage:       "Uso: /remove seguido de una dirección de wallet que empiece por 0x",
	Removed:           "Eliminada %s",
	RemoveMissing:     "Esa dirección no está en tu lista.",
	RemoveFailed:      "No se pudo eliminar. Inténtalo más tarde.",
	ThresholdMoved:    "Ahora el umbral se configura por dirección (1.2 por defecto). Cámbialo en la tarjeta de la dirección en la consola web.",
	NoAddresses:       "Aún no hay direcciones. Añade una con /add 0x...",
	Paused:            "\n\n(Fuera del límite de Starter, alertas en pausa)",

	LoginPrivate: "Confirma el inicio de sesión en un chat privado con el bot.",
	LoginInvalid: "Este enlace de inicio de sesión no es válido o ha caducado. Vuelve a la web y pulsa otra vez «Iniciar sesión con Telegram».",
	LoginPrompt: `Estás iniciando sesión en la consola web de MechBot.

Código que aparece en la web: %s

Pulsa «Confirmar inicio de sesión» solo si los códigos coinciden. Si no acabas de iniciar sesión en la web, no confirmes.`,
	LoginButton:       "Confirmar inicio de sesión",
	LoginExpiredToast: "Enlace caducado",
	LoginExpired:      "Este enlace de inicio de sesión ha caducado. Vuelve a la web y pulsa otra vez «Iniciar sesión con Telegram».",
	LoginFailedToast:  "No se pudo iniciar sesión, inténtalo de nuevo",
	LoginUsedToast:    "El enlace ya no es válido",
	LoginUsed:         "Este enlace de inicio de sesión ya se usó o ha caducado.",
	LoginConfirmToast: "Confirmado",
	LoginConfirmed:    "Inicio de sesión confirmado. Vuelve a la web; la consola se abrirá sola.",

	LangPrompt: "Elige el idioma de los mensajes del bot:",
	LangSet:    "Listo. A partir de ahora los mensajes llegarán en español.",
}

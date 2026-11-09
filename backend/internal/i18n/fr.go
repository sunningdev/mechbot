package i18n

var fr = Text{
	DangerHead: `⚠️ MechBot : alerte de liquidation

Protocole : %s
Adresse : %s
Health factor : %.4f (seuil %.2f)
Collatéral : %s
Dette : %s
`,
	ETHDrop:    "Liquidation si l'ETH baisse encore de %.1f%% (prix de liquidation de l'ETH %s)\n",
	DangerFoot: "\nUne position peut être liquidée dès que son health factor passe sous 1.0.",

	EarlyHead: `⏱ MechBot : alerte anticipée (Pro)

Protocole : %s
Adresse : %s
Health factor on-chain : %.4f
Estimation d'après les prix Binance en direct : %.4f (seuil %.2f)
`,
	EarlyETH:  "ETH : %s → %s (%+.1f%%)\n",
	EarlyNote: "\nLes oracles on-chain ont souvent du retard sur les exchanges : à leur prochaine mise à jour, vous passerez sous votre seuil. Il est encore temps d'agir.\n",

	RateHigh: `📈 MechBot : alerte de taux d'emprunt (Pro)

Protocole : %s
Adresse : %s
Taux d'emprunt %s : %.2f%% par an
Dette : %s, soit environ %s de plus par an à ce taux

Une hausse brutale du taux signifie en général que le pool est presque entièrement emprunté. Votre boucle est peut-être déjà perdante : envisagez de rembourser une partie ou de changer de marché.`,
	BorrowWord: "de la dette",
	RateCalm: `✅ MechBot : le taux d'emprunt est redescendu

Protocole : %s
Adresse : %s
Taux d'emprunt : %.2f%% par an`,

	DepegSideCollateral:   "Collatéral",
	DepegSideDebt:         "Dette",
	DepegEffectCollateral: "Votre collatéral perd de la valeur",
	DepegEffectDebt:       "Votre dette augmente",
	DepegEffectFavorable:  "C'est à votre avantage pour l'instant, mais un décrochage s'accompagne souvent de fortes variations",
	DepegEffectHF:         ", et votre health factor baissera quand les oracles se mettront à jour",
	Depeg: `⚠️ MechBot : alerte de décrochage (Pro)

Protocole : %s
Adresse : %s
%s %s, indexé sur %s, est à %.4f %s (écart de %+.2f%%)
Health factor : %.4f

%s. Pensez à réduire le levier : remboursez une partie du prêt ou ajoutez un autre collatéral.%s`,
	DepegCalm: `✅ MechBot : %s est revenu près de sa parité

Protocole : %s
Adresse : %s
Actuellement %.4f %s`,

	AppLine:     "Agir : %s",
	RescueHead:  "\nQue faire (revenir à %.2f) :\n",
	RescueRepay: "· Rembourser %s\n",
	RescueAdd:   "· Ou ajouter %s de collatéral\n",
	TokenApprox: "%s %s (environ %s)",

	SafeNoDebt: `✅ MechBot : position hors de danger

Protocole : %s
Adresse : %s

Ce prêt est remboursé, il n'y a plus de dette.`,
	Safe: `✅ MechBot : position hors de danger

Protocole : %s
Adresse : %s
Health factor : %.4f (seuil %.2f)
Collatéral : %s
Dette : %s

De nouveau au-dessus de votre seuil.`,

	ListHead:   "Adresse %s\nSeuil %.2f\n",
	ListFailed: "\n%s\nÉchec de la requête : %v\n",
	ListPos:    "\n%s\nHealth factor %.4f\nCollatéral %s\nDette %s\n",
	ListDrop:   "Liquidation si l'ETH baisse de %.1f%% (%s)\n",
	ListAPY:    "Taux d'emprunt %.2f%%\n",
	ListQuiet:  "\nSans dette : %s\n",
	ListSep:    ", ",

	AddedDebt:   "Ajoutée. Positions actuelles :\n\n",
	AddedFailed: "Adresse %s ajoutée.\n\nCertains protocoles n'ont pas répondu pour l'instant. Réessayez /list dans un moment.",
	AddedQuiet: `Adresse %s ajoutée.

Cette adresse n'a aucun prêt en ce moment : nous avons vérifié Aave V3, Compound V3 et Morpho Blue (Ethereum · Arbitrum · Optimism), elle ne peut donc pas être liquidée.

Nous continuons à la surveiller. Si elle emprunte plus tard sur ces protocoles, les alertes démarreront automatiquement, sans besoin de la rajouter.`,
	AddedStarter: "\n\nStarter n'envoie d'alertes que pour Aave V3. Les prêts Compound et Morpho nécessitent Pro.",

	StarterLimits: `· Seules les alertes Aave V3 restent ; celles de Compound et Morpho s'arrêtent
· Les alertes anticipées, de taux et de décrochage s'arrêtent
· La carte des liquidations passe à 1 heure de retard`,
	ProEndingPaid:  "Votre Pro se termine dans moins de 24 heures",
	ProEndingTrial: "Votre essai Pro de 7 jours se termine dans moins de 24 heures",
	ProEnding: `⏳ MechBot : %s

Ensuite, vous passerez à Starter (gratuit) :
· Seules vos %d premières adresses restent surveillées ; les autres sont en pause (pas supprimées)
%s

Garder Pro : %s`,
	ProEndedPaid:   "Pro est terminé, vous êtes passé à Starter",
	ProEndedTrial:  "Votre essai Pro est terminé, vous êtes passé à Starter",
	ProEndedWatch:  "· Vos %d premières adresses restent surveillées",
	ProEndedPaused: " ; les %d autres sont en pause (données conservées, reprise au renouvellement)",
	ProEnded: `MechBot : %s

%s
%s

Renouvelez pour tout retrouver immédiatement : %s`,

	StatusStarter: "Starter (gratuit pour toujours, alertes Aave V3 uniquement)",
	StatusPro:     "Pro (tous les protocoles)",
	StatusTrial:   "Essai Pro (tous les protocoles, retour à Starter à la fin)",
	StatusNone:    "Aucun",
	Status: `Compte MechBot

Offre : %s
Adresses : %d/%d
Pro jusqu'au : %s
Parrain : %s`,

	PlanMonth:   "Pro mensuel",
	PlanQuarter: "Pro trimestriel",
	PlanYear:    "Pro annuel",
	PayTooLow:   "⚠️ %s USDT reçus (%s)\n\nC'est moins que l'offre la moins chère (%d USDT), Pro n'a donc pas été activé. Pour toute question, répondez au bot.",
	PayBonus:    "\n🎁 Bonus de parrainage +%d jours",
	PayDone:     "✅ MechBot Pro activé\n\n%s USDT reçus (%s) : %s, %d jours%s\nPro jusqu'au : %s",
	PayReferrer: "🎁 Récompense de parrainage +%d jours\n\nUne personne que vous avez invitée vient d'acheter %s pour la première fois.\nPro jusqu'au : %s",

	Help: `MechBot : alertes de liquidation

/add 0x... Surveiller une adresse (Starter 3, Pro 20)
/list Health factor par protocole
/remove 0x... Arrêter de surveiller une adresse
/status État du compte
/lang Changer la langue des messages

Chaque adresse démarre avec un seuil de 1.2. Modifiez-le sur la fiche de l'adresse dans la console web.

Starter (gratuit pour toujours) envoie des alertes pour Aave V3 (Ethereum, Arbitrum, Optimism). Pro ajoute Compound V3 et Morpho Blue.`,
	HelpConsole:       "\n\nConsole web : %s",
	TrialNote:         "\n\n🎁 Les nouveaux utilisateurs ont %d jours de Pro : alertes sur tous les protocoles, alertes anticipées et carte des liquidations en direct. Rien n'est facturé à la fin ; vous revenez simplement à l'offre gratuite Starter.",
	UnknownCommand:    "Cette commande n'existe pas.",
	RegisterFailed:    "L'inscription a échoué. Réessayez plus tard.",
	Registered:        "Inscription terminée.",
	AlreadyRegistered: "Vous êtes déjà inscrit.",
	QueryFailed:       "La requête a échoué. Réessayez plus tard.",
	AddUsage:          "Utilisation : /add suivi d'une adresse de wallet commençant par 0x",
	AddFailed:         "Impossible d'ajouter l'adresse. Réessayez plus tard.",
	AddExists:         "Cette adresse est déjà surveillée.",
	AddLimit:          "Votre offre permet jusqu'à %d adresses.",
	RemoveUsage:       "Utilisation : /remove suivi d'une adresse de wallet commençant par 0x",
	Removed:           "Adresse %s supprimée",
	RemoveMissing:     "Cette adresse n'est pas dans votre liste.",
	RemoveFailed:      "Impossible de supprimer l'adresse. Réessayez plus tard.",
	ThresholdMoved:    "Le seuil se règle désormais par adresse (1.2 par défaut). Modifiez-le sur la fiche de l'adresse dans la console web.",
	NoAddresses:       "Aucune adresse pour l'instant. Ajoutez-en une avec /add 0x...",
	Paused:            "\n\n(Au-delà de la limite Starter, alertes en pause)",

	LoginPrivate: "Confirmez la connexion dans une conversation privée avec le bot.",
	LoginInvalid: "Ce lien de connexion est invalide ou a expiré. Retournez sur le site et appuyez à nouveau sur « Se connecter avec Telegram ».",
	LoginPrompt: `Vous vous connectez à la console web de MechBot.

Code affiché sur le site : %s

N'appuyez sur « Confirmer la connexion » que si les codes correspondent. Si vous n'avez pas lancé de connexion sur le site, ne confirmez pas.`,
	LoginButton:       "Confirmer la connexion",
	LoginExpiredToast: "Lien expiré",
	LoginExpired:      "Ce lien de connexion a expiré. Retournez sur le site et appuyez à nouveau sur « Se connecter avec Telegram ».",
	LoginFailedToast:  "Échec de la connexion, réessayez",
	LoginUsedToast:    "Lien plus valide",
	LoginUsed:         "Ce lien de connexion a déjà été utilisé ou a expiré.",
	LoginConfirmToast: "Confirmé",
	LoginConfirmed:    "Connexion confirmée. Retournez sur le site, la console s'ouvrira automatiquement.",

	LangPrompt: "Choisissez la langue des messages du bot :",
	LangSet:    "C'est fait. Les messages seront désormais envoyés en français.",
}

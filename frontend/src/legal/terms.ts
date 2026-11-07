import type { LegalSet } from ".";

export const terms: LegalSet = {
  zh: {
    title: "服务条款",
    summary:
      "简单说：我们只读取你提供的公开地址，用来发送预警；预警不保证送达，也不构成投资建议；付费按套餐开通，如果我们停止服务，会提前通知并按剩余天数退款。",
    updated: "生效日期：2026 年 11 月 7 日",
    body: [
      { type: "h", text: "1. 服务内容" },
      {
        type: "p",
        text: "MechBot 是 DeFi 借贷清算预警工具。你提供钱包地址后，我们读取该地址在 Aave V3、Compound V3、Morpho Blue 等协议上的公开链上数据，在健康因子低于你设置的预警线等情况下，通过 Telegram 私聊提醒你。",
      },
      {
        type: "note",
        text: "MechBot 只读取公开数据：不连接钱包、不需要你签名、不接触你的资产。我们永远不会索要私钥或助记词，任何以 MechBot 名义索要的都是诈骗。",
      },
      { type: "h", text: "2. 账户" },
      {
        type: "list",
        items: [
          "使用 Telegram 登录，一个 Telegram 账号对应一个 MechBot 账户。",
          "请保管好你的 Telegram 账号，通过你的账号进行的操作视为你本人操作。",
          "链上数据是公开的，但请不要用本服务骚扰他人。",
        ],
      },
      { type: "h", text: "3. 套餐与付费" },
      {
        type: "list",
        items: [
          "Starter 免费：最多 3 个地址，仅 Aave V3 预警。",
          "Pro：月付 5 USDT（30 天）、季付 12 USDT（90 天）、年付 36 USDT（365 天）。支持 BNB Chain、Arbitrum、Ethereum、TRON 上的 USDT，TRON 仅支持季付及以上。",
          "到账金额对应的套餐自动开通；超出部分不累计，低于最低套餐价格不开通。请按页面显示的网络和币种转账，转错可能无法找回。",
          "Pro 时长累加：续费和推荐奖励都在现有到期时间上叠加。",
          "新用户可获赠 7 天 Pro 体验，到期自动回到 Starter，不会扣费。",
          "Pro 到期后，超出 Starter 额度的地址会保留但暂停预警，续费后恢复。",
          "价格调整不影响已经购买的时长。",
        ],
      },
      { type: "h", text: "4. 退款与停止服务" },
      {
        type: "list",
        items: [
          "已开通的付费时长一般不退款。转账金额有误等特殊情况，请联系客服处理。",
          "如果我们决定停止服务，会至少提前 30 天通过 Telegram 机器人和网站公告通知，通知期内服务照常运行。",
          "停止服务时，按付费时长的剩余天数折算，以 USDT 退回到你提供的地址。例如年付 36 USDT、剩余 200 天，退还约 19.7 USDT。",
          "赠送的天数（试用、推荐奖励）不参与退款。",
        ],
      },
      { type: "h", text: "5. 预警的局限" },
      {
        type: "p",
        text: "预警依赖区块链节点、价格数据源和 Telegram 等第三方服务，可能延迟、遗漏或失败。我们会尽力保证可靠，例如多节点自动切换、推送失败自动补发，但不保证每条预警都能及时送达。",
      },
      {
        type: "p",
        text: "预警和页面上的数据，包括清算地图、还款和补仓建议，仅供参考，不构成投资建议。你需要自行管理仓位风险。",
      },
      {
        type: "note",
        text: "对于因预警延迟、失败或数据误差造成的任何损失，包括仓位被清算，MechBot 不承担责任。",
      },
      { type: "h", text: "6. 合理使用" },
      {
        type: "list",
        items: [
          "不得攻击、干扰本服务，或绕过套餐限制。",
          "不得批量抓取接口数据或转售本服务。",
          "违反以上规定的账户，我们可以暂停或终止服务。",
        ],
      },
      { type: "h", text: "7. 条款变更" },
      {
        type: "p",
        text: "我们可能更新本条款，更新后会修改本页的生效日期。涉及付费或退款的重大变更，会提前通过 Telegram 通知。变更后继续使用即视为接受。",
      },
      { type: "h", text: "8. 联系我们" },
      { type: "p", text: "有任何问题，请通过首页右下角的 Telegram 客服群，或机器人 @MechFi_bot 联系我们。" },
    ],
  },

  en: {
    title: "Terms of Service",
    summary:
      "In short: we only read the public addresses you give us and send alerts. Alerts are not guaranteed and are not investment advice. Paid plans activate by amount, and if we ever shut down, we give notice and refund the unused days.",
    updated: "Effective: November 7, 2026",
    body: [
      { type: "h", text: "1. The service" },
      {
        type: "p",
        text: "MechBot is a DeFi liquidation alert tool. When you add a wallet address, we read its public on-chain data on protocols such as Aave V3, Compound V3 and Morpho Blue, and message you on Telegram when, for example, a health factor drops below the alert line you set.",
      },
      {
        type: "note",
        text: "MechBot only reads public data. It never connects to your wallet, never asks you to sign anything and never touches your funds. We will never ask for your private key or seed phrase. Anyone who does in our name is a scammer.",
      },
      { type: "h", text: "2. Your account" },
      {
        type: "list",
        items: [
          "You log in with Telegram. One Telegram account is one MechBot account.",
          "Keep your Telegram account safe. Actions taken through it are treated as yours.",
          "On-chain data is public, but do not use the service to harass anyone.",
        ],
      },
      { type: "h", text: "3. Plans and payment" },
      {
        type: "list",
        items: [
          "Starter is free: up to 3 addresses, Aave V3 alerts only.",
          "Pro: 5 USDT monthly (30 days), 12 USDT quarterly (90 days), 36 USDT yearly (365 days). Pay in USDT on BNB Chain, Arbitrum, Ethereum or TRON; TRON accepts quarterly and up.",
          "The plan matching the amount received activates automatically. Any excess is not carried over, and amounts below the cheapest plan do not activate anything. Send only the token and network shown on the page; mistakes may not be recoverable.",
          "Pro time stacks: renewals and referral rewards are added to your current expiry date.",
          "New users get a free 7-day Pro trial. It falls back to Starter at the end, with no charge.",
          "When Pro ends, addresses above the Starter limit are kept but paused, and resume when you renew.",
          "Price changes never affect time you have already bought.",
        ],
      },
      { type: "h", text: "4. Refunds and shutdown" },
      {
        type: "list",
        items: [
          "Paid time is generally non-refundable once activated. For special cases such as a wrong amount, contact support.",
          "If we decide to shut down, we will announce it at least 30 days in advance through the Telegram bot and the website. The service keeps running during that period.",
          "At shutdown, we refund the unused days of your paid time in USDT to an address you provide. Example: a 36 USDT yearly plan with 200 days left gets about 19.7 USDT back.",
          "Free days (trial, referral rewards) are not refunded.",
        ],
      },
      { type: "h", text: "5. Limits of alerts" },
      {
        type: "p",
        text: "Alerts depend on third parties such as blockchain nodes, price feeds and Telegram, and can be delayed, missed or fail. We work hard on reliability, with automatic node failover and retries on failed messages, but we cannot guarantee every alert arrives in time.",
      },
      {
        type: "p",
        text: "Alerts and data on the site, including the liquidation map and repay or top-up suggestions, are for reference only and are not investment advice. You are responsible for managing your positions.",
      },
      {
        type: "note",
        text: "MechBot is not liable for any loss caused by delayed or failed alerts or by data errors, including liquidation of your positions.",
      },
      { type: "h", text: "6. Fair use" },
      {
        type: "list",
        items: [
          "Do not attack or disrupt the service, or work around plan limits.",
          "Do not scrape the API in bulk or resell the service.",
          "We may suspend or close accounts that break these rules.",
        ],
      },
      { type: "h", text: "7. Changes" },
      {
        type: "p",
        text: "We may update these terms and will change the effective date on this page when we do. Material changes to payments or refunds will be announced on Telegram in advance. Continuing to use the service means you accept the changes.",
      },
      { type: "h", text: "8. Contact" },
      { type: "p", text: "Questions? Reach us through the Telegram support group (bottom-right button on the home page) or the bot @MechFi_bot." },
    ],
  },

  ru: {
    title: "Условия использования",
    summary:
      "Коротко: мы только читаем публичные адреса, которые вы добавили, и отправляем оповещения. Доставка оповещений не гарантируется, и это не инвестиционный совет. Платные тарифы включаются по сумме оплаты, а если мы когда-нибудь закроемся, мы заранее предупредим и вернём деньги за неиспользованные дни.",
    updated: "Действует с 7 ноября 2026 г.",
    body: [
      { type: "h", text: "1. Сервис" },
      {
        type: "p",
        text: "MechBot — инструмент оповещений о ликвидации в DeFi. Когда вы добавляете адрес кошелька, мы читаем его публичные ончейн-данные в протоколах Aave V3, Compound V3, Morpho Blue и других и пишем вам в Telegram, например, когда фактор здоровья опускается ниже заданной вами линии оповещения.",
      },
      {
        type: "note",
        text: "MechBot читает только публичные данные: не подключается к кошельку, не просит подписей и не касается ваших средств. Мы никогда не попросим приватный ключ или сид-фразу. Любой, кто просит их от нашего имени, — мошенник.",
      },
      { type: "h", text: "2. Аккаунт" },
      {
        type: "list",
        items: [
          "Вход через Telegram. Один аккаунт Telegram — один аккаунт MechBot.",
          "Берегите свой аккаунт Telegram: действия через него считаются вашими.",
          "Ончейн-данные публичны, но не используйте сервис, чтобы кого-либо преследовать.",
        ],
      },
      { type: "h", text: "3. Тарифы и оплата" },
      {
        type: "list",
        items: [
          "Starter бесплатно: до 3 адресов, оповещения только по Aave V3.",
          "Pro: 5 USDT в месяц (30 дней), 12 USDT в квартал (90 дней), 36 USDT в год (365 дней). Оплата в USDT в сетях BNB Chain, Arbitrum, Ethereum или TRON; в TRON — от квартального тарифа.",
          "Тариф, соответствующий полученной сумме, включается автоматически. Излишек не переносится, сумма ниже минимального тарифа ничего не включает. Отправляйте только указанный на странице токен в указанной сети: ошибку может быть невозможно исправить.",
          "Время Pro суммируется: продления и реферальные бонусы добавляются к текущей дате окончания.",
          "Новые пользователи получают 7 дней Pro бесплатно. После этого аккаунт переходит на Starter без списаний.",
          "Когда Pro заканчивается, адреса сверх лимита Starter сохраняются, но оповещения по ним приостанавливаются до продления.",
          "Изменение цен не затрагивает уже купленное время.",
        ],
      },
      { type: "h", text: "4. Возвраты и закрытие сервиса" },
      {
        type: "list",
        items: [
          "Оплаченное время после включения, как правило, не возвращается. В особых случаях, например при ошибке в сумме, обратитесь в поддержку.",
          "Если мы решим закрыть сервис, мы сообщим об этом минимум за 30 дней через Telegram-бота и сайт. В этот период сервис работает как обычно.",
          "При закрытии мы вернём стоимость неиспользованных оплаченных дней в USDT на указанный вами адрес. Пример: годовой тариф за 36 USDT, осталось 200 дней — возврат около 19,7 USDT.",
          "Бесплатные дни (пробный период, реферальные бонусы) не возвращаются.",
        ],
      },
      { type: "h", text: "5. Ограничения оповещений" },
      {
        type: "p",
        text: "Оповещения зависят от сторонних сервисов — узлов блокчейна, источников цен и Telegram — и могут задерживаться, теряться или не приходить. Мы много делаем для надёжности: автоматическое переключение узлов, повторная отправка при сбоях, — но не можем гарантировать своевременную доставку каждого оповещения.",
      },
      {
        type: "p",
        text: "Оповещения и данные на сайте, включая карту ликвидаций и подсказки, сколько погасить или довнести, носят справочный характер и не являются инвестиционным советом. Вы сами управляете рисками своих позиций.",
      },
      {
        type: "note",
        text: "MechBot не несёт ответственности за убытки из-за задержки или сбоя оповещений либо ошибок в данных, включая ликвидацию ваших позиций.",
      },
      { type: "h", text: "6. Добросовестное использование" },
      {
        type: "list",
        items: [
          "Не атакуйте сервис, не мешайте его работе и не обходите ограничения тарифов.",
          "Не выгружайте данные API массово и не перепродавайте сервис.",
          "Мы можем приостановить или закрыть аккаунты, нарушающие эти правила.",
        ],
      },
      { type: "h", text: "7. Изменения" },
      {
        type: "p",
        text: "Мы можем обновлять эти условия и в этом случае меняем дату вступления в силу на этой странице. О существенных изменениях, касающихся оплаты или возвратов, мы заранее сообщим в Telegram. Продолжая пользоваться сервисом, вы принимаете изменения.",
      },
      { type: "h", text: "8. Контакты" },
      { type: "p", text: "Есть вопросы? Напишите нам в группу поддержки Telegram (кнопка в правом нижнем углу главной страницы) или боту @MechFi_bot." },
    ],
  },

  es: {
    title: "Términos del servicio",
    summary:
      "En resumen: solo leemos las direcciones públicas que nos das y enviamos alertas. Las alertas no están garantizadas y no son asesoramiento de inversión. Los planes de pago se activan según el importe y, si algún día cerramos, avisaremos con antelación y devolveremos los días no usados.",
    updated: "Vigente desde el 7 de noviembre de 2026",
    body: [
      { type: "h", text: "1. El servicio" },
      {
        type: "p",
        text: "MechBot es una herramienta de alertas de liquidación DeFi. Cuando añades una dirección, leemos sus datos públicos on-chain en protocolos como Aave V3, Compound V3 y Morpho Blue, y te escribimos por Telegram cuando, por ejemplo, el factor de salud baja de la línea de alerta que configuraste.",
      },
      {
        type: "note",
        text: "MechBot solo lee datos públicos: nunca se conecta a tu wallet, nunca te pide firmar nada y nunca toca tus fondos. Jamás te pediremos tu clave privada o frase semilla. Quien lo haga en nuestro nombre es un estafador.",
      },
      { type: "h", text: "2. Tu cuenta" },
      {
        type: "list",
        items: [
          "Inicias sesión con Telegram. Una cuenta de Telegram es una cuenta de MechBot.",
          "Protege tu cuenta de Telegram: las acciones hechas con ella se consideran tuyas.",
          "Los datos on-chain son públicos, pero no uses el servicio para acosar a nadie.",
        ],
      },
      { type: "h", text: "3. Planes y pagos" },
      {
        type: "list",
        items: [
          "Starter es gratis: hasta 3 direcciones, solo alertas de Aave V3.",
          "Pro: 5 USDT al mes (30 días), 12 USDT al trimestre (90 días), 36 USDT al año (365 días). Pago en USDT en BNB Chain, Arbitrum, Ethereum o TRON; en TRON, desde el plan trimestral.",
          "El plan que corresponde al importe recibido se activa automáticamente. El exceso no se acumula y un importe inferior al plan más barato no activa nada. Envía solo el token y la red que indica la página: un error puede no tener solución.",
          "El tiempo Pro se acumula: las renovaciones y las recompensas por referidos se suman a tu fecha de vencimiento actual.",
          "Los usuarios nuevos reciben 7 días de Pro gratis. Al terminar vuelven a Starter, sin ningún cargo.",
          "Cuando Pro vence, las direcciones por encima del límite de Starter se conservan pero quedan en pausa, y se reactivan al renovar.",
          "Los cambios de precio nunca afectan al tiempo ya comprado.",
        ],
      },
      { type: "h", text: "4. Reembolsos y cierre" },
      {
        type: "list",
        items: [
          "El tiempo pagado, una vez activado, en general no es reembolsable. Para casos especiales, como un importe equivocado, contacta con soporte.",
          "Si decidimos cerrar el servicio, lo anunciaremos con al menos 30 días de antelación por el bot de Telegram y la web. Durante ese periodo el servicio sigue funcionando.",
          "Al cerrar, devolveremos los días pagados no usados en USDT a la dirección que nos indiques. Ejemplo: un plan anual de 36 USDT con 200 días restantes recibe unos 19,7 USDT.",
          "Los días gratuitos (prueba, recompensas por referidos) no se reembolsan.",
        ],
      },
      { type: "h", text: "5. Límites de las alertas" },
      {
        type: "p",
        text: "Las alertas dependen de terceros como nodos de blockchain, fuentes de precios y Telegram, y pueden retrasarse, perderse o fallar. Trabajamos mucho en la fiabilidad, con cambio automático de nodo y reintentos de mensajes fallidos, pero no podemos garantizar que cada alerta llegue a tiempo.",
      },
      {
        type: "p",
        text: "Las alertas y los datos de la web, incluido el mapa de liquidaciones y las sugerencias de cuánto pagar o añadir, son solo orientativos y no constituyen asesoramiento de inversión. Tú eres responsable de gestionar tus posiciones.",
      },
      {
        type: "note",
        text: "MechBot no se hace responsable de ninguna pérdida causada por alertas retrasadas o fallidas o por errores en los datos, incluida la liquidación de tus posiciones.",
      },
      { type: "h", text: "6. Uso razonable" },
      {
        type: "list",
        items: [
          "No ataques ni interfieras con el servicio, ni eludas los límites de los planes.",
          "No extraigas datos de la API de forma masiva ni revendas el servicio.",
          "Podemos suspender o cerrar las cuentas que incumplan estas normas.",
        ],
      },
      { type: "h", text: "7. Cambios" },
      {
        type: "p",
        text: "Podemos actualizar estos términos y, cuando lo hagamos, cambiaremos la fecha de vigencia de esta página. Los cambios importantes sobre pagos o reembolsos se anunciarán antes por Telegram. Seguir usando el servicio implica aceptar los cambios.",
      },
      { type: "h", text: "8. Contacto" },
      { type: "p", text: "¿Dudas? Escríbenos en el grupo de soporte de Telegram (botón abajo a la derecha en la página de inicio) o al bot @MechFi_bot." },
    ],
  },

  fr: {
    title: "Conditions d'utilisation",
    summary:
      "En bref : nous lisons uniquement les adresses publiques que vous nous donnez et envoyons des alertes. Les alertes ne sont pas garanties et ne constituent pas un conseil en investissement. Les offres payantes s'activent selon le montant reçu et, si nous fermons un jour, nous prévenons à l'avance et remboursons les jours non utilisés.",
    updated: "En vigueur depuis le 7 novembre 2026",
    body: [
      { type: "h", text: "1. Le service" },
      {
        type: "p",
        text: "MechBot est un outil d'alerte de liquidation DeFi. Quand vous ajoutez une adresse, nous lisons ses données publiques on-chain sur des protocoles comme Aave V3, Compound V3 et Morpho Blue, et vous écrivons sur Telegram quand, par exemple, le facteur de santé passe sous le seuil d'alerte que vous avez choisi.",
      },
      {
        type: "note",
        text: "MechBot lit uniquement des données publiques : il ne se connecte jamais à votre wallet, ne vous demande jamais de signer et ne touche jamais à vos fonds. Nous ne vous demanderons jamais votre clé privée ni votre phrase de récupération. Toute personne qui le fait en notre nom est un escroc.",
      },
      { type: "h", text: "2. Votre compte" },
      {
        type: "list",
        items: [
          "Vous vous connectez avec Telegram. Un compte Telegram correspond à un compte MechBot.",
          "Protégez votre compte Telegram : les actions effectuées avec lui sont considérées comme les vôtres.",
          "Les données on-chain sont publiques, mais n'utilisez pas le service pour harceler qui que ce soit.",
        ],
      },
      { type: "h", text: "3. Offres et paiement" },
      {
        type: "list",
        items: [
          "Starter est gratuit : jusqu'à 3 adresses, alertes Aave V3 uniquement.",
          "Pro : 5 USDT par mois (30 jours), 12 USDT par trimestre (90 jours), 36 USDT par an (365 jours). Paiement en USDT sur BNB Chain, Arbitrum, Ethereum ou TRON ; sur TRON, à partir de l'offre trimestrielle.",
          "L'offre correspondant au montant reçu s'active automatiquement. L'excédent n'est pas reporté et un montant inférieur à l'offre la moins chère n'active rien. N'envoyez que le jeton et le réseau indiqués sur la page : une erreur peut être irrécupérable.",
          "Le temps Pro se cumule : les renouvellements et les récompenses de parrainage s'ajoutent à votre date d'expiration actuelle.",
          "Les nouveaux utilisateurs reçoivent 7 jours de Pro gratuits. À la fin, le compte repasse en Starter, sans aucun prélèvement.",
          "Quand Pro expire, les adresses au-delà de la limite Starter sont conservées mais mises en pause, puis reprennent au renouvellement.",
          "Un changement de prix n'affecte jamais le temps déjà acheté.",
        ],
      },
      { type: "h", text: "4. Remboursements et fermeture" },
      {
        type: "list",
        items: [
          "Le temps payé n'est en général pas remboursable une fois activé. Pour les cas particuliers, comme un montant erroné, contactez le support.",
          "Si nous décidons de fermer le service, nous l'annoncerons au moins 30 jours à l'avance via le bot Telegram et le site. Le service continue de fonctionner pendant cette période.",
          "À la fermeture, nous remboursons les jours payés non utilisés en USDT à l'adresse que vous indiquez. Exemple : une offre annuelle à 36 USDT avec 200 jours restants donne droit à environ 19,7 USDT.",
          "Les jours offerts (essai, récompenses de parrainage) ne sont pas remboursés.",
        ],
      },
      { type: "h", text: "5. Limites des alertes" },
      {
        type: "p",
        text: "Les alertes dépendent de tiers comme les nœuds blockchain, les sources de prix et Telegram, et peuvent être retardées, manquées ou échouer. Nous travaillons dur sur la fiabilité, avec bascule automatique entre nœuds et renvoi des messages échoués, mais nous ne pouvons pas garantir que chaque alerte arrive à temps.",
      },
      {
        type: "p",
        text: "Les alertes et les données du site, y compris la carte des liquidations et les suggestions de remboursement ou de collatéral à ajouter, sont fournies à titre indicatif et ne constituent pas un conseil en investissement. Vous êtes responsable de la gestion de vos positions.",
      },
      {
        type: "note",
        text: "MechBot n'est pas responsable des pertes causées par des alertes retardées ou échouées ou par des erreurs de données, y compris la liquidation de vos positions.",
      },
      { type: "h", text: "6. Usage raisonnable" },
      {
        type: "list",
        items: [
          "N'attaquez pas et ne perturbez pas le service, et ne contournez pas les limites des offres.",
          "N'aspirez pas les données de l'API en masse et ne revendez pas le service.",
          "Nous pouvons suspendre ou fermer les comptes qui enfreignent ces règles.",
        ],
      },
      { type: "h", text: "7. Modifications" },
      {
        type: "p",
        text: "Nous pouvons mettre à jour ces conditions et modifierons alors la date d'entrée en vigueur sur cette page. Les changements importants concernant les paiements ou les remboursements seront annoncés à l'avance sur Telegram. Continuer à utiliser le service vaut acceptation des modifications.",
      },
      { type: "h", text: "8. Contact" },
      { type: "p", text: "Une question ? Contactez-nous via le groupe de support Telegram (bouton en bas à droite de la page d'accueil) ou le bot @MechFi_bot." },
    ],
  },

  pt: {
    title: "Termos de serviço",
    summary:
      "Resumindo: só lemos os endereços públicos que você nos passa e enviamos alertas. Os alertas não são garantidos e não são recomendação de investimento. Os planos pagos são ativados conforme o valor recebido e, se um dia encerrarmos, avisaremos com antecedência e devolveremos os dias não usados.",
    updated: "Em vigor desde 7 de novembro de 2026",
    body: [
      { type: "h", text: "1. O serviço" },
      {
        type: "p",
        text: "MechBot é uma ferramenta de alertas de liquidação DeFi. Quando você adiciona um endereço, lemos os dados públicos on-chain dele em protocolos como Aave V3, Compound V3 e Morpho Blue e te avisamos no Telegram quando, por exemplo, o fator de saúde cai abaixo da linha de alerta que você definiu.",
      },
      {
        type: "note",
        text: "MechBot só lê dados públicos: nunca se conecta à sua carteira, nunca pede assinaturas e nunca toca nos seus fundos. Jamais pediremos sua chave privada ou frase-semente. Quem pedir em nosso nome é golpista.",
      },
      { type: "h", text: "2. Sua conta" },
      {
        type: "list",
        items: [
          "O login é feito com o Telegram. Uma conta do Telegram é uma conta do MechBot.",
          "Proteja sua conta do Telegram: ações feitas por ela são consideradas suas.",
          "Os dados on-chain são públicos, mas não use o serviço para assediar ninguém.",
        ],
      },
      { type: "h", text: "3. Planos e pagamento" },
      {
        type: "list",
        items: [
          "Starter é grátis: até 3 endereços, alertas apenas da Aave V3.",
          "Pro: 5 USDT por mês (30 dias), 12 USDT por trimestre (90 dias), 36 USDT por ano (365 dias). Pagamento em USDT na BNB Chain, Arbitrum, Ethereum ou TRON; na TRON, a partir do plano trimestral.",
          "O plano correspondente ao valor recebido é ativado automaticamente. O excedente não é acumulado e valores abaixo do plano mais barato não ativam nada. Envie apenas o token e a rede indicados na página: um erro pode não ter volta.",
          "O tempo Pro se acumula: renovações e recompensas de indicação são somadas à sua data de vencimento atual.",
          "Novos usuários ganham 7 dias de Pro grátis. Ao final, a conta volta para o Starter, sem nenhuma cobrança.",
          "Quando o Pro vence, os endereços acima do limite do Starter são mantidos, mas ficam pausados, e voltam a funcionar quando você renova.",
          "Mudanças de preço nunca afetam o tempo já comprado.",
        ],
      },
      { type: "h", text: "4. Reembolsos e encerramento" },
      {
        type: "list",
        items: [
          "O tempo pago, depois de ativado, em geral não é reembolsável. Para casos especiais, como valor errado, fale com o suporte.",
          "Se decidirmos encerrar o serviço, avisaremos com pelo menos 30 dias de antecedência pelo bot do Telegram e pelo site. Nesse período o serviço continua funcionando.",
          "No encerramento, devolvemos os dias pagos não usados em USDT para o endereço que você informar. Exemplo: um plano anual de 36 USDT com 200 dias restantes recebe cerca de 19,7 USDT.",
          "Dias gratuitos (teste, recompensas de indicação) não são reembolsados.",
        ],
      },
      { type: "h", text: "5. Limites dos alertas" },
      {
        type: "p",
        text: "Os alertas dependem de terceiros, como nós de blockchain, fontes de preço e o Telegram, e podem atrasar, se perder ou falhar. Trabalhamos muito pela confiabilidade, com troca automática de nó e reenvio de mensagens que falharam, mas não podemos garantir que todo alerta chegue a tempo.",
      },
      {
        type: "p",
        text: "Os alertas e os dados do site, incluindo o mapa de liquidações e as sugestões de quanto pagar ou adicionar, servem apenas como referência e não são recomendação de investimento. Você é responsável por gerenciar suas posições.",
      },
      {
        type: "note",
        text: "O MechBot não se responsabiliza por perdas causadas por alertas atrasados ou que falharam, ou por erros nos dados, incluindo a liquidação das suas posições.",
      },
      { type: "h", text: "6. Uso justo" },
      {
        type: "list",
        items: [
          "Não ataque nem atrapalhe o serviço, e não contorne os limites dos planos.",
          "Não extraia dados da API em massa nem revenda o serviço.",
          "Podemos suspender ou encerrar contas que violem essas regras.",
        ],
      },
      { type: "h", text: "7. Alterações" },
      {
        type: "p",
        text: "Podemos atualizar estes termos e, quando isso acontecer, mudaremos a data de vigência nesta página. Mudanças importantes sobre pagamentos ou reembolsos serão anunciadas antes no Telegram. Continuar usando o serviço significa aceitar as mudanças.",
      },
      { type: "h", text: "8. Contato" },
      { type: "p", text: "Dúvidas? Fale com a gente no grupo de suporte do Telegram (botão no canto inferior direito da página inicial) ou pelo bot @MechFi_bot." },
    ],
  },

  vi: {
    title: "Điều khoản dịch vụ",
    summary:
      "Tóm tắt: chúng tôi chỉ đọc các địa chỉ công khai bạn cung cấp và gửi cảnh báo. Cảnh báo không được bảo đảm và không phải lời khuyên đầu tư. Gói trả phí được kích hoạt theo số tiền nhận được, và nếu một ngày chúng tôi ngừng dịch vụ, chúng tôi sẽ báo trước và hoàn tiền cho số ngày chưa dùng.",
    updated: "Có hiệu lực từ ngày 7 tháng 11 năm 2026",
    body: [
      { type: "h", text: "1. Dịch vụ" },
      {
        type: "p",
        text: "MechBot là công cụ cảnh báo thanh lý DeFi. Khi bạn thêm một địa chỉ ví, chúng tôi đọc dữ liệu on-chain công khai của địa chỉ đó trên các giao thức như Aave V3, Compound V3 và Morpho Blue, và nhắn bạn trên Telegram khi, chẳng hạn, hệ số sức khỏe giảm dưới ngưỡng cảnh báo bạn đặt.",
      },
      {
        type: "note",
        text: "MechBot chỉ đọc dữ liệu công khai: không kết nối ví, không yêu cầu ký và không chạm vào tài sản của bạn. Chúng tôi sẽ không bao giờ hỏi khóa riêng hay cụm từ khôi phục. Bất kỳ ai hỏi những thứ đó dưới danh nghĩa MechBot đều là lừa đảo.",
      },
      { type: "h", text: "2. Tài khoản" },
      {
        type: "list",
        items: [
          "Đăng nhập bằng Telegram. Một tài khoản Telegram là một tài khoản MechBot.",
          "Hãy giữ an toàn tài khoản Telegram: mọi thao tác qua tài khoản đó được xem là của bạn.",
          "Dữ liệu on-chain là công khai, nhưng đừng dùng dịch vụ để quấy rối người khác.",
        ],
      },
      { type: "h", text: "3. Gói dịch vụ và thanh toán" },
      {
        type: "list",
        items: [
          "Starter miễn phí: tối đa 3 địa chỉ, chỉ cảnh báo Aave V3.",
          "Pro: 5 USDT mỗi tháng (30 ngày), 12 USDT mỗi quý (90 ngày), 36 USDT mỗi năm (365 ngày). Thanh toán bằng USDT trên BNB Chain, Arbitrum, Ethereum hoặc TRON; TRON chỉ nhận từ gói quý trở lên.",
          "Gói tương ứng với số tiền nhận được sẽ tự động kích hoạt. Phần dư không được cộng dồn, số tiền thấp hơn gói rẻ nhất sẽ không kích hoạt gì. Chỉ gửi đúng token và mạng hiển thị trên trang: gửi nhầm có thể không lấy lại được.",
          "Thời gian Pro được cộng dồn: gia hạn và thưởng giới thiệu đều cộng vào ngày hết hạn hiện tại.",
          "Người dùng mới được tặng 7 ngày Pro. Hết hạn sẽ tự quay về Starter, không bị trừ tiền.",
          "Khi Pro hết hạn, các địa chỉ vượt quá giới hạn Starter vẫn được giữ nhưng tạm dừng cảnh báo, và hoạt động lại khi bạn gia hạn.",
          "Thay đổi giá không ảnh hưởng đến thời gian bạn đã mua.",
        ],
      },
      { type: "h", text: "4. Hoàn tiền và ngừng dịch vụ" },
      {
        type: "list",
        items: [
          "Thời gian đã trả phí sau khi kích hoạt nhìn chung không được hoàn tiền. Với trường hợp đặc biệt, như chuyển sai số tiền, hãy liên hệ hỗ trợ.",
          "Nếu quyết định ngừng dịch vụ, chúng tôi sẽ thông báo trước ít nhất 30 ngày qua bot Telegram và trang web. Trong thời gian đó dịch vụ vẫn chạy bình thường.",
          "Khi ngừng dịch vụ, chúng tôi hoàn lại số ngày trả phí chưa dùng bằng USDT về địa chỉ bạn cung cấp. Ví dụ: gói năm 36 USDT còn 200 ngày sẽ được hoàn khoảng 19,7 USDT.",
          "Các ngày được tặng (dùng thử, thưởng giới thiệu) không được hoàn tiền.",
        ],
      },
      { type: "h", text: "5. Giới hạn của cảnh báo" },
      {
        type: "p",
        text: "Cảnh báo phụ thuộc vào bên thứ ba như node blockchain, nguồn giá và Telegram, nên có thể bị chậm, bỏ sót hoặc thất bại. Chúng tôi nỗ lực đảm bảo độ tin cậy, với tự động chuyển node và gửi lại tin nhắn bị lỗi, nhưng không thể bảo đảm mọi cảnh báo đều đến kịp thời.",
      },
      {
        type: "p",
        text: "Cảnh báo và dữ liệu trên trang, bao gồm bản đồ thanh lý và gợi ý cần trả bao nhiêu hoặc bổ sung bao nhiêu, chỉ mang tính tham khảo và không phải lời khuyên đầu tư. Bạn tự chịu trách nhiệm quản lý vị thế của mình.",
      },
      {
        type: "note",
        text: "MechBot không chịu trách nhiệm cho bất kỳ tổn thất nào do cảnh báo chậm, thất bại hoặc sai lệch dữ liệu, bao gồm việc vị thế của bạn bị thanh lý.",
      },
      { type: "h", text: "6. Sử dụng hợp lý" },
      {
        type: "list",
        items: [
          "Không tấn công, gây gián đoạn dịch vụ hoặc lách giới hạn của gói.",
          "Không thu thập hàng loạt dữ liệu API hoặc bán lại dịch vụ.",
          "Chúng tôi có thể tạm khóa hoặc chấm dứt tài khoản vi phạm các quy định này.",
        ],
      },
      { type: "h", text: "7. Thay đổi điều khoản" },
      {
        type: "p",
        text: "Chúng tôi có thể cập nhật điều khoản này và sẽ đổi ngày hiệu lực trên trang này khi cập nhật. Các thay đổi quan trọng về thanh toán hoặc hoàn tiền sẽ được thông báo trước qua Telegram. Tiếp tục sử dụng dịch vụ đồng nghĩa với việc bạn chấp nhận thay đổi.",
      },
      { type: "h", text: "8. Liên hệ" },
      { type: "p", text: "Có thắc mắc? Liên hệ chúng tôi qua nhóm hỗ trợ Telegram (nút ở góc dưới bên phải trang chủ) hoặc bot @MechFi_bot." },
    ],
  },
};

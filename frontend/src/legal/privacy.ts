import type { LegalSet } from ".";

export const privacy: LegalSet = {
  zh: {
    title: "隐私政策",
    summary: "我们只收集提供预警必需的最少数据：不要邮箱、不要密码，不出售任何数据。",
    updated: "生效日期：2026 年 11 月 7 日",
    body: [
      { type: "h", text: "1. 我们收集什么" },
      {
        type: "list",
        items: [
          "Telegram 账户信息：数字 ID、用户名和名字，用于登录和发送预警。",
          "你添加的钱包地址和预警线设置。",
          "付款记录：付款网络、金额、交易哈希、套餐和到期时间，以及为你分配的专属充值地址。",
          "推荐关系：你是通过谁的邀请链接注册的。",
        ],
      },
      { type: "h", text: "2. 我们不收集什么" },
      {
        type: "list",
        items: [
          "不要邮箱、手机号和密码。",
          "不要私钥、助记词，也不连接你的钱包。",
          "MechBot 程序不记录你的 IP 地址，也不使用广告或第三方统计追踪。网页服务器出于安全和排障目的，可能短期保留访问日志，日志不与你的账户关联使用。",
        ],
      },
      { type: "h", text: "3. 数据如何使用" },
      {
        type: "list",
        items: [
          "读取地址的链上仓位，判断是否需要发预警。",
          "通过 Telegram 给你发送预警和服务通知，例如付款到账、Pro 即将到期。",
          "开通和计算 Pro 时长、推荐奖励，以及停止服务时计算退款。",
        ],
      },
      { type: "p", text: "我们不会出售、出租你的数据，也不会把它用于广告。" },
      { type: "h", text: "4. 第三方服务" },
      { type: "p", text: "为了提供服务，部分请求会经过以下第三方：" },
      {
        type: "list",
        items: [
          "Telegram：登录和消息推送。",
          "区块链节点服务商，以及 Etherscan、Morpho 的数据接口：查询地址的公开链上数据，这些服务商能看到被查询的地址。",
          "币安：获取公开行情价格。首页的实时价格条由你的浏览器直接连接币安的公开行情接口。",
        ],
      },
      { type: "h", text: "5. Cookie 和本地存储" },
      {
        type: "list",
        items: [
          "一个登录 Cookie，只用于保持登录状态，页面上的 JavaScript 无法读取。",
          "浏览器本地存储：你选择的语言，以及邀请链接里的推荐人 ID。",
        ],
      },
      { type: "h", text: "6. 数据保存与删除" },
      {
        type: "list",
        items: [
          "你在控制台删除地址后，该地址和它的预警记录会从我们的数据库删除。",
          "如需删除账户和全部数据，请联系客服。付款记录会保留，用于对账和退款计算。",
          "区块链上的数据是公开且无法删除的，不在我们的控制范围内。",
        ],
      },
      { type: "h", text: "7. 数据安全" },
      {
        type: "list",
        items: [
          "全站 HTTPS 加密传输。",
          "收款私钥离线保存，服务器上只有收款公钥，碰不到任何资金。",
          "服务器访问受限，配置文件只有管理员可读。",
        ],
      },
      { type: "h", text: "8. 政策变更与联系" },
      {
        type: "p",
        text: "我们可能更新本政策，更新后会修改本页的生效日期。有任何问题，请通过首页右下角的 Telegram 客服群，或机器人 @MechFi_bot 联系我们。",
      },
    ],
  },

  en: {
    title: "Privacy Policy",
    summary: "We collect only the minimum needed to send alerts: no email, no password, and we never sell data.",
    updated: "Effective: November 7, 2026",
    body: [
      { type: "h", text: "1. What we collect" },
      {
        type: "list",
        items: [
          "Telegram account details: numeric ID, username and first name, used to log you in and send alerts.",
          "The wallet addresses you add and their alert line settings.",
          "Payment records: network, amount, transaction hash, plan and expiry date, plus the deposit address assigned to you.",
          "Referrals: whose invite link you signed up with.",
        ],
      },
      { type: "h", text: "2. What we don't collect" },
      {
        type: "list",
        items: [
          "No email, phone number or password.",
          "No private key or seed phrase, and no wallet connection.",
          "The MechBot app does not record your IP address and uses no ads or third-party analytics. The web server may keep short-lived access logs for security and troubleshooting; they are not used together with your account.",
        ],
      },
      { type: "h", text: "3. How we use it" },
      {
        type: "list",
        items: [
          "Reading the on-chain positions of your addresses to decide when to alert.",
          "Sending alerts and service notices on Telegram, such as payment received or Pro ending soon.",
          "Activating and counting Pro time and referral rewards, and calculating refunds if we ever shut down.",
        ],
      },
      { type: "p", text: "We never sell or rent your data, and we never use it for advertising." },
      { type: "h", text: "4. Third parties" },
      { type: "p", text: "To run the service, some requests go through these third parties:" },
      {
        type: "list",
        items: [
          "Telegram: login and message delivery.",
          "Blockchain node providers and the Etherscan and Morpho data APIs: reading public on-chain data for your addresses. These providers can see the addresses being queried.",
          "Binance: public market prices. The live price ticker on the home page connects from your browser directly to Binance's public market data stream.",
        ],
      },
      { type: "h", text: "5. Cookies and local storage" },
      {
        type: "list",
        items: [
          "One login cookie, used only to keep you signed in. Page scripts cannot read it.",
          "Browser local storage: your chosen language and the referrer ID from an invite link.",
        ],
      },
      { type: "h", text: "6. Retention and deletion" },
      {
        type: "list",
        items: [
          "When you delete an address in the console, it and its alert history are deleted from our database.",
          "To delete your account and all its data, contact support. Payment records are kept for accounting and refund calculations.",
          "Data on the blockchain is public and permanent, and outside our control.",
        ],
      },
      { type: "h", text: "7. Security" },
      {
        type: "list",
        items: [
          "HTTPS everywhere.",
          "Payment private keys are kept offline. The server only holds public keys and cannot touch any funds.",
          "Server access is restricted, and configuration files are readable by administrators only.",
        ],
      },
      { type: "h", text: "8. Changes and contact" },
      {
        type: "p",
        text: "We may update this policy and will change the effective date on this page when we do. Questions? Reach us through the Telegram support group (bottom-right button on the home page) or the bot @MechFi_bot.",
      },
    ],
  },

  ru: {
    title: "Политика конфиденциальности",
    summary: "Мы собираем только минимум, нужный для оповещений: без почты и паролей, и никогда не продаём данные.",
    updated: "Действует с 7 ноября 2026 г.",
    body: [
      { type: "h", text: "1. Что мы собираем" },
      {
        type: "list",
        items: [
          "Данные аккаунта Telegram: числовой ID, имя пользователя и имя — для входа и отправки оповещений.",
          "Добавленные вами адреса кошельков и настройки линий оповещения.",
          "Платежи: сеть, сумма, хеш транзакции, тариф и дата окончания, а также выделенный вам адрес для оплаты.",
          "Рефералы: по чьей пригласительной ссылке вы зарегистрировались.",
        ],
      },
      { type: "h", text: "2. Что мы не собираем" },
      {
        type: "list",
        items: [
          "Никакой почты, номера телефона и пароля.",
          "Никаких приватных ключей и сид-фраз, никакого подключения кошелька.",
          "Приложение MechBot не записывает ваш IP-адрес и не использует рекламу или сторонние системы аналитики. Веб-сервер может недолго хранить журналы доступа для безопасности и диагностики; они не используются вместе с данными вашего аккаунта.",
        ],
      },
      { type: "h", text: "3. Как мы используем данные" },
      {
        type: "list",
        items: [
          "Читаем ончейн-позиции ваших адресов, чтобы решить, когда отправить оповещение.",
          "Отправляем оповещения и служебные уведомления в Telegram, например о поступлении оплаты или скором окончании Pro.",
          "Включаем и считаем время Pro и реферальные бонусы, а при закрытии сервиса — суммы возврата.",
        ],
      },
      { type: "p", text: "Мы никогда не продаём и не сдаём в аренду ваши данные и не используем их для рекламы." },
      { type: "h", text: "4. Сторонние сервисы" },
      { type: "p", text: "Для работы сервиса часть запросов проходит через следующих третьих лиц:" },
      {
        type: "list",
        items: [
          "Telegram: вход и доставка сообщений.",
          "Провайдеры узлов блокчейна и API данных Etherscan и Morpho: чтение публичных ончейн-данных ваших адресов. Эти провайдеры видят запрашиваемые адреса.",
          "Binance: публичные рыночные цены. Бегущая строка цен на главной странице подключается из вашего браузера напрямую к публичному потоку рыночных данных Binance.",
        ],
      },
      { type: "h", text: "5. Cookie и локальное хранилище" },
      {
        type: "list",
        items: [
          "Один cookie для входа — только чтобы вы оставались в аккаунте. Скрипты страницы не могут его прочитать.",
          "Локальное хранилище браузера: выбранный язык и ID пригласившего из пригласительной ссылки.",
        ],
      },
      { type: "h", text: "6. Хранение и удаление" },
      {
        type: "list",
        items: [
          "Когда вы удаляете адрес в консоли, он и история оповещений по нему удаляются из нашей базы данных.",
          "Чтобы удалить аккаунт и все его данные, обратитесь в поддержку. Записи о платежах сохраняются для учёта и расчёта возвратов.",
          "Данные в блокчейне публичны и неизменяемы и находятся вне нашего контроля.",
        ],
      },
      { type: "h", text: "7. Безопасность" },
      {
        type: "list",
        items: [
          "HTTPS на всём сайте.",
          "Приватные ключи для приёма платежей хранятся офлайн. На сервере только публичные ключи, он не может распоряжаться средствами.",
          "Доступ к серверу ограничен, конфигурационные файлы доступны только администраторам.",
        ],
      },
      { type: "h", text: "8. Изменения и контакты" },
      {
        type: "p",
        text: "Мы можем обновлять эту политику и в этом случае меняем дату вступления в силу на этой странице. Есть вопросы? Напишите нам в группу поддержки Telegram (кнопка в правом нижнем углу главной страницы) или боту @MechFi_bot.",
      },
    ],
  },

  es: {
    title: "Política de privacidad",
    summary: "Solo recopilamos lo mínimo para enviarte alertas: sin email, sin contraseña, y nunca vendemos datos.",
    updated: "Vigente desde el 7 de noviembre de 2026",
    body: [
      { type: "h", text: "1. Qué recopilamos" },
      {
        type: "list",
        items: [
          "Datos de tu cuenta de Telegram: ID numérico, nombre de usuario y nombre, para iniciar sesión y enviarte alertas.",
          "Las direcciones que añades y su línea de alerta.",
          "Pagos: red, importe, hash de la transacción, plan y fecha de vencimiento, además de la dirección de depósito que te asignamos.",
          "Referidos: con el enlace de invitación de quién te registraste.",
        ],
      },
      { type: "h", text: "2. Qué no recopilamos" },
      {
        type: "list",
        items: [
          "Ni email, ni teléfono, ni contraseña.",
          "Ni clave privada ni frase semilla, y ninguna conexión con tu wallet.",
          "La aplicación MechBot no registra tu dirección IP ni usa publicidad o analítica de terceros. El servidor web puede guardar registros de acceso durante poco tiempo por seguridad y para resolver problemas; no se usan junto con tu cuenta.",
        ],
      },
      { type: "h", text: "3. Cómo usamos los datos" },
      {
        type: "list",
        items: [
          "Leer las posiciones on-chain de tus direcciones para decidir cuándo avisarte.",
          "Enviarte alertas y avisos del servicio por Telegram, como pago recibido o Pro a punto de vencer.",
          "Activar y contar el tiempo Pro y las recompensas por referidos, y calcular reembolsos si algún día cerramos.",
        ],
      },
      { type: "p", text: "Nunca vendemos ni alquilamos tus datos, ni los usamos para publicidad." },
      { type: "h", text: "4. Servicios de terceros" },
      { type: "p", text: "Para funcionar, algunas solicitudes pasan por estos terceros:" },
      {
        type: "list",
        items: [
          "Telegram: inicio de sesión y entrega de mensajes.",
          "Proveedores de nodos de blockchain y las API de datos de Etherscan y Morpho: lectura de datos públicos on-chain de tus direcciones. Estos proveedores pueden ver las direcciones consultadas.",
          "Binance: precios públicos de mercado. La barra de precios en vivo de la página de inicio se conecta desde tu navegador directamente al flujo público de datos de mercado de Binance.",
        ],
      },
      { type: "h", text: "5. Cookies y almacenamiento local" },
      {
        type: "list",
        items: [
          "Una cookie de sesión, solo para mantenerte conectado. Los scripts de la página no pueden leerla.",
          "Almacenamiento local del navegador: el idioma elegido y el ID de quien te invitó, si llegaste con un enlace de invitación.",
        ],
      },
      { type: "h", text: "6. Conservación y eliminación" },
      {
        type: "list",
        items: [
          "Cuando eliminas una dirección en la consola, se borran de nuestra base de datos la dirección y su historial de alertas.",
          "Para eliminar tu cuenta y todos sus datos, contacta con soporte. Los registros de pago se conservan para la contabilidad y el cálculo de reembolsos.",
          "Los datos de la blockchain son públicos y permanentes, y están fuera de nuestro control.",
        ],
      },
      { type: "h", text: "7. Seguridad" },
      {
        type: "list",
        items: [
          "HTTPS en toda la web.",
          "Las claves privadas de cobro se guardan offline. El servidor solo tiene claves públicas y no puede tocar ningún fondo.",
          "El acceso al servidor está restringido y los archivos de configuración solo los pueden leer los administradores.",
        ],
      },
      { type: "h", text: "8. Cambios y contacto" },
      {
        type: "p",
        text: "Podemos actualizar esta política y, cuando lo hagamos, cambiaremos la fecha de vigencia de esta página. ¿Dudas? Escríbenos en el grupo de soporte de Telegram (botón abajo a la derecha en la página de inicio) o al bot @MechFi_bot.",
      },
    ],
  },

  fr: {
    title: "Politique de confidentialité",
    summary: "Nous collectons uniquement le minimum nécessaire aux alertes : ni e-mail ni mot de passe, et nous ne vendons jamais de données.",
    updated: "En vigueur depuis le 7 novembre 2026",
    body: [
      { type: "h", text: "1. Ce que nous collectons" },
      {
        type: "list",
        items: [
          "Les informations de votre compte Telegram : identifiant numérique, nom d'utilisateur et prénom, pour vous connecter et vous envoyer des alertes.",
          "Les adresses que vous ajoutez et leur seuil d'alerte.",
          "Les paiements : réseau, montant, hash de transaction, offre et date d'expiration, ainsi que l'adresse de dépôt qui vous est attribuée.",
          "Le parrainage : le lien d'invitation avec lequel vous vous êtes inscrit.",
        ],
      },
      { type: "h", text: "2. Ce que nous ne collectons pas" },
      {
        type: "list",
        items: [
          "Ni e-mail, ni numéro de téléphone, ni mot de passe.",
          "Ni clé privée, ni phrase de récupération, et aucune connexion à votre wallet.",
          "L'application MechBot n'enregistre pas votre adresse IP et n'utilise ni publicité ni outil d'analyse tiers. Le serveur web peut conserver brièvement des journaux d'accès pour la sécurité et le dépannage ; ils ne sont pas utilisés avec votre compte.",
        ],
      },
      { type: "h", text: "3. Comment nous les utilisons" },
      {
        type: "list",
        items: [
          "Lire les positions on-chain de vos adresses pour décider quand vous alerter.",
          "Vous envoyer des alertes et des notifications de service sur Telegram, par exemple paiement reçu ou Pro bientôt expiré.",
          "Activer et décompter le temps Pro et les récompenses de parrainage, et calculer les remboursements si nous fermons un jour.",
        ],
      },
      { type: "p", text: "Nous ne vendons ni ne louons jamais vos données et ne les utilisons jamais à des fins publicitaires." },
      { type: "h", text: "4. Services tiers" },
      { type: "p", text: "Pour faire fonctionner le service, certaines requêtes passent par ces tiers :" },
      {
        type: "list",
        items: [
          "Telegram : connexion et envoi des messages.",
          "Les fournisseurs de nœuds blockchain et les API de données d'Etherscan et de Morpho : lecture des données publiques on-chain de vos adresses. Ces fournisseurs peuvent voir les adresses interrogées.",
          "Binance : prix publics du marché. Le bandeau de prix en direct de la page d'accueil se connecte depuis votre navigateur directement au flux public de données de marché de Binance.",
        ],
      },
      { type: "h", text: "5. Cookies et stockage local" },
      {
        type: "list",
        items: [
          "Un cookie de connexion, utilisé uniquement pour vous garder connecté. Les scripts de la page ne peuvent pas le lire.",
          "Le stockage local du navigateur : la langue choisie et l'identifiant du parrain issu d'un lien d'invitation.",
        ],
      },
      { type: "h", text: "6. Conservation et suppression" },
      {
        type: "list",
        items: [
          "Quand vous supprimez une adresse dans la console, l'adresse et son historique d'alertes sont supprimés de notre base de données.",
          "Pour supprimer votre compte et toutes ses données, contactez le support. Les enregistrements de paiement sont conservés pour la comptabilité et le calcul des remboursements.",
          "Les données de la blockchain sont publiques et permanentes, et échappent à notre contrôle.",
        ],
      },
      { type: "h", text: "7. Sécurité" },
      {
        type: "list",
        items: [
          "HTTPS sur tout le site.",
          "Les clés privées d'encaissement sont conservées hors ligne. Le serveur ne détient que des clés publiques et ne peut toucher à aucun fonds.",
          "L'accès au serveur est restreint et les fichiers de configuration ne sont lisibles que par les administrateurs.",
        ],
      },
      { type: "h", text: "8. Modifications et contact" },
      {
        type: "p",
        text: "Nous pouvons mettre à jour cette politique et modifierons alors la date d'entrée en vigueur sur cette page. Une question ? Contactez-nous via le groupe de support Telegram (bouton en bas à droite de la page d'accueil) ou le bot @MechFi_bot.",
      },
    ],
  },

  pt: {
    title: "Política de privacidade",
    summary: "Coletamos só o mínimo necessário para enviar alertas: sem e-mail, sem senha, e nunca vendemos dados.",
    updated: "Em vigor desde 7 de novembro de 2026",
    body: [
      { type: "h", text: "1. O que coletamos" },
      {
        type: "list",
        items: [
          "Dados da sua conta do Telegram: ID numérico, nome de usuário e nome, para fazer login e enviar alertas.",
          "Os endereços que você adiciona e a linha de alerta de cada um.",
          "Pagamentos: rede, valor, hash da transação, plano e data de vencimento, além do endereço de depósito atribuído a você.",
          "Indicações: com o link de convite de quem você se cadastrou.",
        ],
      },
      { type: "h", text: "2. O que não coletamos" },
      {
        type: "list",
        items: [
          "Nada de e-mail, telefone ou senha.",
          "Nada de chave privada ou frase-semente, e nenhuma conexão com a sua carteira.",
          "O aplicativo MechBot não registra seu endereço IP e não usa anúncios nem ferramentas de análise de terceiros. O servidor web pode manter registros de acesso por pouco tempo, por segurança e para resolver problemas; eles não são usados junto com a sua conta.",
        ],
      },
      { type: "h", text: "3. Como usamos os dados" },
      {
        type: "list",
        items: [
          "Ler as posições on-chain dos seus endereços para decidir quando avisar.",
          "Enviar alertas e avisos do serviço pelo Telegram, como pagamento recebido ou Pro perto de vencer.",
          "Ativar e contar o tempo Pro e as recompensas de indicação, e calcular reembolsos se um dia encerrarmos.",
        ],
      },
      { type: "p", text: "Nunca vendemos nem alugamos seus dados, e nunca os usamos para publicidade." },
      { type: "h", text: "4. Serviços de terceiros" },
      { type: "p", text: "Para funcionar, algumas solicitações passam por estes terceiros:" },
      {
        type: "list",
        items: [
          "Telegram: login e entrega de mensagens.",
          "Provedores de nós de blockchain e as APIs de dados da Etherscan e da Morpho: leitura dos dados públicos on-chain dos seus endereços. Esses provedores podem ver os endereços consultados.",
          "Binance: preços públicos de mercado. A faixa de preços ao vivo da página inicial se conecta do seu navegador diretamente ao fluxo público de dados de mercado da Binance.",
        ],
      },
      { type: "h", text: "5. Cookies e armazenamento local" },
      {
        type: "list",
        items: [
          "Um cookie de login, usado só para manter você conectado. Os scripts da página não conseguem lê-lo.",
          "Armazenamento local do navegador: o idioma escolhido e o ID de quem te indicou, vindo de um link de convite.",
        ],
      },
      { type: "h", text: "6. Retenção e exclusão" },
      {
        type: "list",
        items: [
          "Quando você exclui um endereço no console, ele e o histórico de alertas dele são apagados do nosso banco de dados.",
          "Para excluir sua conta e todos os dados, fale com o suporte. Os registros de pagamento são mantidos para contabilidade e cálculo de reembolsos.",
          "Os dados na blockchain são públicos e permanentes, e estão fora do nosso controle.",
        ],
      },
      { type: "h", text: "7. Segurança" },
      {
        type: "list",
        items: [
          "HTTPS em todo o site.",
          "As chaves privadas de recebimento ficam offline. O servidor só tem chaves públicas e não consegue mexer em nenhum fundo.",
          "O acesso ao servidor é restrito e os arquivos de configuração só podem ser lidos por administradores.",
        ],
      },
      { type: "h", text: "8. Alterações e contato" },
      {
        type: "p",
        text: "Podemos atualizar esta política e, quando isso acontecer, mudaremos a data de vigência nesta página. Dúvidas? Fale com a gente no grupo de suporte do Telegram (botão no canto inferior direito da página inicial) ou pelo bot @MechFi_bot.",
      },
    ],
  },

  vi: {
    title: "Chính sách quyền riêng tư",
    summary: "Chúng tôi chỉ thu thập dữ liệu tối thiểu cần để gửi cảnh báo: không email, không mật khẩu, và không bao giờ bán dữ liệu.",
    updated: "Có hiệu lực từ ngày 7 tháng 11 năm 2026",
    body: [
      { type: "h", text: "1. Chúng tôi thu thập gì" },
      {
        type: "list",
        items: [
          "Thông tin tài khoản Telegram: ID số, tên người dùng và tên, dùng để đăng nhập và gửi cảnh báo.",
          "Các địa chỉ ví bạn thêm và ngưỡng cảnh báo của từng địa chỉ.",
          "Lịch sử thanh toán: mạng, số tiền, mã giao dịch, gói và ngày hết hạn, cùng địa chỉ nạp tiền riêng được cấp cho bạn.",
          "Giới thiệu: bạn đăng ký qua liên kết mời của ai.",
        ],
      },
      { type: "h", text: "2. Chúng tôi không thu thập gì" },
      {
        type: "list",
        items: [
          "Không email, số điện thoại hay mật khẩu.",
          "Không khóa riêng, không cụm từ khôi phục, và không kết nối ví.",
          "Ứng dụng MechBot không ghi lại địa chỉ IP của bạn và không dùng quảng cáo hay công cụ phân tích của bên thứ ba. Máy chủ web có thể lưu nhật ký truy cập trong thời gian ngắn vì mục đích bảo mật và xử lý sự cố; nhật ký này không được dùng kèm với tài khoản của bạn.",
        ],
      },
      { type: "h", text: "3. Chúng tôi dùng dữ liệu thế nào" },
      {
        type: "list",
        items: [
          "Đọc vị thế on-chain của các địa chỉ để quyết định khi nào cần cảnh báo.",
          "Gửi cảnh báo và thông báo dịch vụ qua Telegram, ví dụ đã nhận thanh toán hoặc Pro sắp hết hạn.",
          "Kích hoạt và tính thời gian Pro, thưởng giới thiệu, và tính tiền hoàn lại nếu một ngày chúng tôi ngừng dịch vụ.",
        ],
      },
      { type: "p", text: "Chúng tôi không bao giờ bán hay cho thuê dữ liệu của bạn, và không dùng nó cho quảng cáo." },
      { type: "h", text: "4. Dịch vụ bên thứ ba" },
      { type: "p", text: "Để vận hành dịch vụ, một số yêu cầu đi qua các bên thứ ba sau:" },
      {
        type: "list",
        items: [
          "Telegram: đăng nhập và gửi tin nhắn.",
          "Nhà cung cấp node blockchain và API dữ liệu của Etherscan, Morpho: đọc dữ liệu on-chain công khai của các địa chỉ. Các nhà cung cấp này có thể thấy địa chỉ được truy vấn.",
          "Binance: giá thị trường công khai. Thanh giá trực tiếp trên trang chủ được trình duyệt của bạn kết nối thẳng tới luồng dữ liệu thị trường công khai của Binance.",
        ],
      },
      { type: "h", text: "5. Cookie và bộ nhớ cục bộ" },
      {
        type: "list",
        items: [
          "Một cookie đăng nhập, chỉ để giữ bạn ở trạng thái đăng nhập. Script trên trang không đọc được cookie này.",
          "Bộ nhớ cục bộ của trình duyệt: ngôn ngữ bạn chọn và ID người giới thiệu từ liên kết mời.",
        ],
      },
      { type: "h", text: "6. Lưu trữ và xóa dữ liệu" },
      {
        type: "list",
        items: [
          "Khi bạn xóa một địa chỉ trong bảng điều khiển, địa chỉ đó và lịch sử cảnh báo của nó bị xóa khỏi cơ sở dữ liệu của chúng tôi.",
          "Để xóa tài khoản và toàn bộ dữ liệu, hãy liên hệ hỗ trợ. Lịch sử thanh toán được giữ lại để đối soát và tính tiền hoàn lại.",
          "Dữ liệu trên blockchain là công khai, vĩnh viễn và nằm ngoài tầm kiểm soát của chúng tôi.",
        ],
      },
      { type: "h", text: "7. Bảo mật" },
      {
        type: "list",
        items: [
          "Toàn bộ trang web dùng HTTPS.",
          "Khóa riêng nhận thanh toán được lưu ngoại tuyến. Máy chủ chỉ có khóa công khai và không thể chạm vào bất kỳ khoản tiền nào.",
          "Quyền truy cập máy chủ bị giới hạn, tệp cấu hình chỉ quản trị viên mới đọc được.",
        ],
      },
      { type: "h", text: "8. Thay đổi và liên hệ" },
      {
        type: "p",
        text: "Chúng tôi có thể cập nhật chính sách này và sẽ đổi ngày hiệu lực trên trang này khi cập nhật. Có thắc mắc? Liên hệ chúng tôi qua nhóm hỗ trợ Telegram (nút ở góc dưới bên phải trang chủ) hoặc bot @MechFi_bot.",
      },
    ],
  },
};

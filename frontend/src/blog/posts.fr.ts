import type { Post } from "./posts";

export const frPosts: Post[] = [
  {
    slug: "health-factor",
    title: "Comment se calcule le facteur de santé ? Comprendre le chiffre « avant liquidation »",
    summary:
      "Aave, Compound et Morpho nomment leur seuil de liquidation différemment, mais c'est la même idée. Une formule l'explique et vous dit de combien les prix peuvent encore baisser.",
    date: "2026-10-28",
    tag: "Fondamentaux",
    readMin: 5,
    body: [
      {
        type: "p",
        text: "Un protocole de prêt ne se soucie pas de combien vous avez emprunté. Une seule chose compte : une fois votre collatéral décoté selon son seuil de liquidation, couvre-t-il encore votre dette ? Le facteur de santé (Health Factor, HF) est ce ratio.",
      },
      { type: "h", text: "Une formule" },
      { type: "code", text: "HF = Σ(valeur du collatéral × seuil de liquidation) / valeur de la dette" },
      {
        type: "p",
        text: "Au-dessus de 1, tout va bien. À 1, vous atteignez le seuil de liquidation. En dessous de 1, n'importe qui peut liquider votre position et prendre une partie de votre collatéral en bonus.",
      },
      { type: "h", text: "Trois protocoles, trois noms" },
      {
        type: "list",
        items: [
          "Aave V3 : chaque actif a un Liquidation Threshold, et le contrat renvoie directement le facteur de santé.",
          "Compound V3 : chaque actif de collatéral a un liquidateCollateralFactor ; nous appliquons nous-mêmes la même formule.",
          "Morpho Blue : chaque marché a un LLTV (ratio prêt/valeur de liquidation). Même formule, avec un seul actif de collatéral.",
        ],
      },
      { type: "h", text: "Ce que signifie « avant liquidation »" },
      {
        type: "p",
        text: "Dans la console, « avant liquidation » indique de combien le prix de votre collatéral peut baisser, à dette constante, avant que la liquidation ne se déclenche :",
      },
      { type: "code", text: "Avant liquidation = 1 − 1 / HF" },
      {
        type: "p",
        text: "Avec un HF de 1.25, votre collatéral peut encore perdre 20 %. Avec un HF de 1.07, il ne reste qu'environ 6.5 % : une seule grosse bougie rouge suffit.",
      },
      {
        type: "note",
        text: "Si vous empruntez un actif volatil (par exemple de l'ETH), votre dette bouge aussi, donc la marge réelle sera différente. Voyez ce chiffre comme un repère intuitif, pas comme une prédiction exacte.",
      },
      { type: "h", text: "Où placer votre seuil d'alerte" },
      {
        type: "p",
        text: "Avec des stablecoins en collatéral pour emprunter des stablecoins, vous pouvez descendre vers 1.1. Avec de l'ETH ou du BTC en collatéral pour emprunter des stablecoins, 1.3 à 1.5 laisse le temps de recharger.",
      },
    ],
  },
  {
    slug: "rpc-failover",
    title: "Que se passe-t-il quand un nœud tombe ? La bascule multi-nœuds de MechBot",
    summary:
      "Les RPC publics limitent les requêtes, expirent et tombent en panne. Voici comment les alertes arrivent malgré tout quand les nœuds lâchent, et pourquoi « illisible » n'est jamais traité comme « sûr ».",
    date: "2026-10-26",
    tag: "Ingénierie",
    readMin: 6,
    body: [
      {
        type: "p",
        text: "La pire panne pour une alerte de liquidation n'est pas une erreur, c'est le silence : un nœud tombe, rien n'est lu, rien n'est envoyé, et l'utilisateur pense que tout va bien. La bascule de MechBot repose sur une règle : une donnée qu'on n'a pas pu lire n'est jamais considérée comme sûre.",
      },
      { type: "h", text: "Chaque chaîne a un groupe de nœuds" },
      {
        type: "p",
        text: "Chaque chaîne dispose de plusieurs points d'accès RPC : d'abord ceux que vous configurez, puis des réserves publiques intégrées. Chaque requête les essaie dans l'ordre, retient celui qui a fonctionné et commence par lui la fois suivante.",
      },
      { type: "code", text: "Nœud configuré → réserve 1 → réserve 2 → …\nSuccès : retenir ce nœud\nÉchec : passer au suivant" },
      { type: "h", text: "Quand tous les nœuds échouent" },
      {
        type: "list",
        items: [
          "La position est marquée « requête échouée, nouvel essai au prochain tour » dans la console, jamais affichée comme sûre.",
          "Si une source de données échoue pour toutes les adresses plusieurs tours de suite, une alerte d'astreinte part vers le Telegram de l'opérateur.",
          "Une fois le nœud rétabli, un avis de rétablissement confirme que le chemin est de nouveau sain.",
        ],
      },
      { type: "h", text: "Le programme doit tenir, lui aussi" },
      {
        type: "p",
        text: "La lecture des protocoles, la boucle de vérification, le scan des paiements et le traitement des messages du bot ont chacun leur propre protection contre les plantages. Une donnée erronée sur une position ne peut pas faire échouer tout un tour, encore moins arrêter le service.",
      },
      {
        type: "note",
        text: "Quand l'API Telegram renvoie une erreur, le jeton du bot est retiré du journal afin que les secrets ne fuient jamais dans les systèmes de logs.",
      },
    ],
  },
  {
    slug: "no-spam",
    title: "Pourquoi MechBot ne spamme pas : la vie d'une alerte",
    summary:
      "Des centaines de notifications par jour, c'est comme aucune. Nous ne vous prévenons que lorsque vous franchissez votre seuil, et jamais deux fois en une heure pour la même position, sauf si elle est sur le point d'être liquidée.",
    date: "2026-10-24",
    tag: "Produit",
    readMin: 4,
    body: [
      {
        type: "p",
        text: "Beaucoup de bots de surveillance envoient chaque variation de prix. Les utilisateurs les coupent, puis ratent la seule alerte qui compte. MechBot vise l'inverse : silencieux par défaut, bruyant seulement avant la liquidation.",
      },
      { type: "h", text: "Vérifier : toutes les 5 minutes" },
      {
        type: "p",
        text: "Toutes les 5 minutes, le serveur lit les positions de toutes vos adresses sur chaque protocole, directement dans les contrats on-chain, sans API de données tierce.",
      },
      { type: "h", text: "Décider : un seul seuil" },
      {
        type: "list",
        items: [
          "Pas de dette, ou facteur de santé au-dessus de votre seuil : rien n'est envoyé.",
          "Passage sous le seuil : message immédiat avec protocole, chaîne, facteur de santé, collatéral et dette.",
          "Reste sous le seuil : pas de répétition pendant 1 heure ; si rien ne s'est arrangé après une heure, un rappel de plus.",
          "Facteur de santé sous 1.05, au bord de la liquidation : un rappel toutes les 15 minutes jusqu'à ce que vous agissiez ou que la position se rétablisse.",
          "Retour au-dessus du seuil : un seul message « de nouveau en sécurité », puis le silence.",
        ],
      },
      { type: "code", text: "Sûr ──baisse──▶ Alerte (message)\n  ▲             │ rappel toutes les heures (15 min sous 1.05)\n  └──reprise── De nouveau en sécurité (un message)" },
      { type: "h", text: "Le résultat" },
      {
        type: "p",
        text: "Une position qui passe de sûre à dangereuse puis revient ne vous coûte en général que deux messages. Chacun mérite que vous interrompiez ce que vous faites pour y jeter un œil.",
      },
    ],
  },
];

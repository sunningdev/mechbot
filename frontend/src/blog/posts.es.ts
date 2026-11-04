import type { Post } from "./posts";

export const esPosts: Post[] = [
  {
    slug: "health-factor",
    title: "¿Cómo se calcula el factor de salud? Entender la cifra “hasta liquidación”",
    summary:
      "Aave, Compound y Morpho llaman de forma distinta a su umbral de liquidación, pero es la misma idea. Una fórmula lo explica y te dice cuánto pueden caer todavía los precios.",
    date: "2026-10-28",
    tag: "Fundamentos",
    readMin: 5,
    body: [
      {
        type: "p",
        text: "A un protocolo de préstamos no le importa cuánto pediste. Le importa una sola cosa: después de descontar tu colateral según su umbral de liquidación, ¿sigue cubriendo tu deuda? El factor de salud (Health Factor, HF) es ese cociente.",
      },
      { type: "h", text: "Una fórmula" },
      { type: "code", text: "HF = Σ(valor del colateral × umbral de liquidación) / valor de la deuda" },
      {
        type: "p",
        text: "Por encima de 1 es seguro. En 1 llegas al umbral de liquidación. Por debajo de 1, cualquiera puede liquidar tu posición y quedarse parte de tu colateral como bonificación.",
      },
      { type: "h", text: "Tres protocolos, tres nombres" },
      {
        type: "list",
        items: [
          "Aave V3: cada activo tiene un Liquidation Threshold y el contrato devuelve directamente el factor de salud.",
          "Compound V3: cada activo de colateral tiene un liquidateCollateralFactor; aplicamos nosotros la misma fórmula.",
          "Morpho Blue: cada mercado tiene un LLTV (préstamo sobre valor de liquidación). Misma fórmula, con un único activo de colateral.",
        ],
      },
      { type: "h", text: "Qué significa “hasta liquidación”" },
      {
        type: "p",
        text: "En la consola, “hasta liquidación” es cuánto puede bajar el precio de tu colateral, con la deuda sin cambios, antes de que se active la liquidación:",
      },
      { type: "code", text: "Hasta liquidación = 1 − 1 / HF" },
      {
        type: "p",
        text: "Con HF 1.25 tu colateral aún puede caer un 20%. Con HF 1.07 solo queda alrededor de un 6.5%: basta una gran vela roja.",
      },
      {
        type: "note",
        text: "Si pides prestado un activo volátil (por ejemplo, ETH), tu deuda también se mueve, así que el margen real será distinto. Toma esta cifra como una guía intuitiva, no como una predicción exacta.",
      },
      { type: "h", text: "Dónde poner tu umbral de alerta" },
      {
        type: "p",
        text: "Con colateral en stablecoins y deuda en stablecoins puedes ir más bajo, alrededor de 1.1. Con colateral en ETH o BTC y deuda en stablecoins, entre 1.3 y 1.5 es un buen rango que te deja tiempo para reponer.",
      },
    ],
  },
  {
    slug: "rpc-failover",
    title: "¿Qué pasa cuando cae un nodo? La conmutación multinodo de MechBot",
    summary:
      "Los RPC públicos limitan peticiones, se quedan sin respuesta y se caen. Así seguimos entregando alertas cuando fallan los nodos, y por qué “no se pudo leer” nunca se trata como “seguro”.",
    date: "2026-10-26",
    tag: "Ingeniería",
    readMin: 6,
    body: [
      {
        type: "p",
        text: "El peor fallo de una alerta de liquidación no es un error, es el silencio: un nodo cae, no se lee nada, no se envía nada y el usuario cree que todo va bien. La conmutación de MechBot se basa en una regla: los datos que no pudimos leer nunca se consideran seguros.",
      },
      { type: "h", text: "Cada cadena tiene un grupo de nodos" },
      {
        type: "p",
        text: "Cada cadena tiene varios endpoints RPC: primero los que configuras tú y después reservas públicas integradas. Cada consulta los prueba en orden, recuerda el que funcionó y empieza por él la próxima vez.",
      },
      { type: "code", text: "Nodo configurado → reserva 1 → reserva 2 → …\nÉxito: recordar este nodo\nFallo: pasar al siguiente" },
      { type: "h", text: "Cuando fallan todos los nodos" },
      {
        type: "list",
        items: [
          "La posición se marca como “consulta fallida, se reintentará en la próxima ronda” en la consola; nunca se muestra como segura.",
          "Si una fuente de datos falla para todas las direcciones varias rondas seguidas, se envía una alerta de guardia al Telegram del operador.",
          "Cuando el nodo se recupera, un aviso de recuperación confirma que la ruta vuelve a estar sana.",
        ],
      },
      { type: "h", text: "El programa también tiene que aguantar" },
      {
        type: "p",
        text: "Las lecturas de protocolos, el sondeo, el escaneo de pagos y el manejo de mensajes del bot tienen cada uno su propia protección contra caídas. Los datos erróneos de una posición no pueden tumbar una ronda completa, y menos aún el servicio.",
      },
      {
        type: "note",
        text: "Cuando la API de Telegram devuelve un error, el token del bot se elimina del log para que los secretos nunca acaben en los sistemas de registro.",
      },
    ],
  },
  {
    slug: "no-spam",
    title: "Por qué MechBot no hace spam: la vida de una alerta",
    summary:
      "Cientos de notificaciones al día equivalen a ninguna. Solo te avisamos cuando cruzas tu umbral de alerta, y nunca dos veces en una hora por la misma posición.",
    date: "2026-10-24",
    tag: "Producto",
    readMin: 4,
    body: [
      {
        type: "p",
        text: "Muchos bots de monitorización envían cada movimiento del precio. Los usuarios los silencian y luego se pierden la única alerta que importa. MechBot busca lo contrario: silencio por defecto y un aviso fuerte solo antes de la liquidación.",
      },
      { type: "h", text: "Comprobar: cada 5 minutos" },
      {
        type: "p",
        text: "Cada 5 minutos el backend lee las posiciones de todas tus direcciones en cada protocolo, directamente de los contratos on-chain, sin APIs de datos de terceros.",
      },
      { type: "h", text: "Decidir: un solo umbral" },
      {
        type: "list",
        items: [
          "Sin deuda, o factor de salud por encima de tu umbral: no se envía nada.",
          "Cae por debajo del umbral: mensaje inmediato con protocolo, cadena, factor de salud, colateral y deuda.",
          "Sigue por debajo: no se repite en 1 hora; si pasada una hora aún no se ha recuperado, un recordatorio más.",
          "Vuelve por encima del umbral: un único mensaje de “de vuelta a salvo” y de nuevo silencio.",
        ],
      },
      { type: "code", text: "Seguro ──cae──▶ Alerta (mensaje)\n  ▲              │ silencio 1 hora\n  └──se recupera── De vuelta a salvo (un mensaje)" },
      { type: "h", text: "El resultado" },
      {
        type: "p",
        text: "Que una posición pase de segura a peligrosa y vuelva suele costarte solo dos mensajes. Cada uno merece que dejes lo que estás haciendo para mirarlo.",
      },
    ],
  },
];

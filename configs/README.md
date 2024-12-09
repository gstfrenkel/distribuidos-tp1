# Explicación de las configuraciones de los workers

Tomando por ejemplo:

```json
{
  "query": 5,
  "peers": 0,
  "expected-eofs": 2,
  "input-queues": [
    {
      "name": "top_queue"
    }
  ],
  "output-queues": [
    {
      "exchange": "reports",
      "name": "reports_%d",
      "key": "%d",
      "consumers": 1
    }
  ],
  "exchanges": [
    {
      "name": "reports",
      "kind": "direct"
    }
  ],
  "log-level": "INFO"
}
```

- `query`: Este campo es custom. Refiere a datos particulares que necesite el nodo para funcionar. Por ejemplo, en el json que vemos arriba, representa el N del topN.
- `peers`: Cantidad de nodos del mismo tipo que existen.
- `expected-eofs`:  cantidad de EOFs que un nodo espera recibir antes de propagar información. Se usa para diferenciar un aggregator de un worker "normal".
- `input-queues`: Lista de colas de las que el nodo va a consumir mensajes.
  - `name`: Nombre de la cola.
  - `exchange`: Nombre del exchange al que está asociada la cola (sólo si `peers` > 0). Se usa para reencolar EOFs.
  - `key`: Clave de enrutamiento de la cola (sólo si `peers` > 0). Se usa para reencolar EOFs.
- `output-queues`: Lista de colas a las que el nodo va a enviar mensajes.
  - `name`: Nombre de la cola.
  - `exchange`: Nombre del exchange al que está asociada la cola.
  - `key`: Clave de enrutamiento de la cola.
  - `consumers`: Cantidad de consumidores que va a tener la cola.
  - `single`: De ser true, define si un output del tipo XXX_%d tiene q explotarse en N outputs (XXX_0, XXX_1, ...) o uno solo (XXX_0). Por ejemplo:
    - no single: sería para alimentar joiners
    - single: sería para alimentar un worker no joiner
- `exchanges`: Lista de exchanges que el nodo va a declarar.
  - `name`: Nombre del exchange.
  - `kind`: Tipo de exchange.
- `log-level`: Nivel de loggeo del nodo.

## ¿Qué atributos debería modificar de escalar un nodo?
Se deben ajustar las configuraciones de los nodos del tipo escalado y de los adyacentes. En particular los campos `peers`, `consumers` y/o `expected-eofs` según el caso.

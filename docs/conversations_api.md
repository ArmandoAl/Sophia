# Conversations API

Fecha: 2026-07-14

Todos los endpoints requieren `Authorization: Bearer <token>`.

## POST /conversations

Request:

```json
{
  "title": "optional"
}
```

Response `201`:

```json
{
  "conversation": {
    "id": "...",
    "user_id": "...",
    "title": "optional",
    "status": "active",
    "created_at": "...",
    "updated_at": "..."
  }
}
```

## GET /conversations

Query opcional:

- `status`
- `limit`
- `cursor`

Response `200`:

```json
{
  "conversations": [],
  "next_cursor": "..."
}
```

## GET /conversations/{id}

Response `200`:

```json
{
  "conversation": {}
}
```

`404` tambien cubre conversaciones de otro usuario.

## GET /conversations/{id}/messages

Query opcional:

- `limit`
- `cursor`

Response `200`, orden cronologico ascendente:

```json
{
  "messages": [],
  "next_cursor": "..."
}
```

## POST /conversations/{id}/messages

Request:

```json
{
  "content": "Hola Sofia"
}
```

Response `201`:

```json
{
  "conversation": {},
  "user_message": {
    "role": "user",
    "content": "Hola Sofia"
  },
  "assistant_message": {
    "role": "assistant",
    "content": "..."
  },
  "proposed_actions": [],
  "runtime_request_id": "..."
}
```

El backend invoca AI Runtime con `dry_run=true`; por lo tanto no persiste proposals ni ejecuta tools.

Si falla el provider/runtime, el mensaje user queda guardado y no se crea respuesta assistant falsa.

## POST /conversations/{id}/archive

Response `200`:

```json
{
  "conversation": {
    "status": "archived"
  }
}
```

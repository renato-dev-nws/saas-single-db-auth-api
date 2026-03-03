# Upload de imagens com working-progress via SSE

## Contexto

Após o upload, as imagens são processadas de forma assíncrona pelo `worker-images` (converte para WebP e gera variantes medium/small/thumb). O backend agora expõe um endpoint SSE que notifica quando o processamento termina.

**Endpoint SSE:**
```
GET /api/:url_code/images/:image_id/events
```
Eventos emitidos:
- `pending` — conectado, aguardando processamento
- `completed` — processamento concluído, `data` contém o objeto completo da imagem com `original_url`, `medium_url`, `small_url`, `thumb_url`
- `timeout` — 90s sem resposta do worker
- `error` — erro ao carregar imagem

---

## O que mudar na função `uploadImages`

A função atual faz upload e já chama `fetchImages()` — mas nesse momento as variantes ainda não estão prontas (status `pending`).

A nova lógica:
1. Faz o upload normalmente (retorna `{ images: [{ image_id, public_url }] }`)
2. Para cada `image_id` retornado, abre uma conexão SSE
3. Ao receber `completed`, chama `fetchImages()` para recarregar a lista

### Tipo da resposta de upload

```ts
interface UploadedImage {
  image_id: string
  public_url: string
}
```

### Função atualizada

```ts
async function uploadImages() {
  if (!selectedFiles.value.length) return
  uploading.value = true

  try {
    const formData = new FormData()
    for (const file of selectedFiles.value) {
      formData.append('images', file)
    }

    // 1. Upload — resposta imediata com os arquivos originais
    const response = await apiFetch<{ images: UploadedImage[] }>(
      `/${props.urlCode}/${props.entityType}/${props.entityId}/images`,
      { method: 'POST', body: formData }
    )

    toast.add({
      severity: 'success',
      summary: t('common.success'),
      detail: t('images.imagesUploaded'),
      life: 3000,
    })

    clearUpload()

    // 2. Recarrega imediatamente para mostrar os originais (status pending)
    fetchImages()

    // 3. Abre SSE para cada imagem — ao completar, recarrega novamente
    for (const img of response.images) {
      watchImageProcessing(img.image_id)
    }
  } catch (error: any) {
    handleError(error)
  } finally {
    uploading.value = false
  }
}

function watchImageProcessing(imageId: string) {
  // Monta a URL base igual ao apiFetch usa, sem o /api prefix se necessário
  const baseURL = useRuntimeConfig().public.apiBase // ex: http://localhost:8080/api

  // Passa o token como query param porque EventSource não suporta headers
  const token = useCookie('auth_token').value // ajuste para onde seu token está armazenado
  const url = `${baseURL}/${props.urlCode}/images/${imageId}/events?token=${token}`

  const es = new EventSource(url)

  es.addEventListener('completed', () => {
    fetchImages() // recarrega a lista com as variantes prontas
    es.close()
  })

  es.addEventListener('timeout', () => {
    es.close()
  })

  es.addEventListener('error', () => {
    es.close()
  })

  // Segurança: fecha após 95s de qualquer forma
  setTimeout(() => es.close(), 95_000)
}
```

---

## Atenção: auth no SSE

`EventSource` **não suporta headers customizados**. Se o endpoint SSE exigir autenticação, há duas opções:

**Opção A — query param `?token=...` (mais simples)**

Passar o JWT como query string e validar no backend no middleware do Gin.

**Opção B — cookie HttpOnly**

Se o token já for enviado via cookie automaticamente pelo browser, o `EventSource` o envia junto e não é necessária nenhuma mudança.

Consulte como o seu `apiFetch` está autenticado para saber qual opção usar.

---

## Comportamento visual esperado

| Momento | O que o usuário vê |
|---|---|
| Imediatamente após upload | Imagens aparecem na lista (versão original, status `pending`) |
| Worker termina (≈ 2–5s) | Lista recarrega automaticamente com versões WebP + variantes |
| Server timeout (90s) | Nada muda visualmente — lista fica como está |

Opcionalmente, se quiser mostrar um spinner sobre as imagens em processamento, filtre as imagens da lista pelo campo `processing_status === 'pending'` e sobreponha um indicador visual.

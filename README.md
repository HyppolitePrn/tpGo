# taskrunner

Orchestrateur de tâches concurrentes en Go (stdlib uniquement). Lit une liste de tâches depuis un fichier JSON, les exécute en parallèle avec un pool de workers, gère timeout/retries par tâche, puis produit un rapport JSON et un fichier `METRICS.md`.

## Lancer le programme

```bash
make build            # compile le binaire ./taskrunner
make run              # build + exécute avec tasks.json
make test             # lance les tests unitaires (-race)
make lint             # go vet + gofmt -d
```

Ou directement :

```bash
go run ./cmd/taskrunner -file tasks.json -workers 3 -verbose
```

Flags :
- `-file` : chemin du fichier JSON de tâches (obligatoire)
- `-workers` : nombre de workers concurrents, entre 1 et 100 (défaut : 3)
- `-verbose` : affiche en temps réel le statut de chaque tâche sur stderr

Le rapport JSON est écrit sur **stdout**, `METRICS.md` est généré à la racine du dossier courant. `Ctrl+C` déclenche un arrêt propre : les tâches en cours sont annulées et un rapport partiel est quand même produit.

## Format de `tasks.json`

```json
{
  "tasks": [
    { "id": "t1", "type": "print", "params": { "message": "hello" }, "timeout": "2s", "retries": 0 }
  ]
}
```

- `id` : identifiant unique de la tâche
- `type` : `print`, `calc`, `download` ou `fake`
- `params` : paramètres spécifiques au type
- `timeout` : durée max d'exécution (format Go, ex. `"2s"`, `"500ms"`)
- `retries` : nombre de tentatives supplémentaires en cas d'échec ou de timeout

### Types de tâches

| Type | Params | Comportement |
|---|---|---|
| `print` | `message` | Affiche le message sur stdout |
| `calc` | `value` (nombre) | Calcule le carré de `value` |
| `download` | `url`, `dest` | Télécharge `url` en pur Go (`net/http`) vers `dest` |
| `fake` | `behavior` (`success`/`fail`/`timeout`), `delay` | Simule une tâche, utile pour les tests |

Chaque tâche est relancée automatiquement (jusqu'à `retries` fois) si elle échoue ou dépasse son `timeout`.

## Rapport de sortie

```json
{
  "results": [
    { "id": "t1", "status": "success", "duration": "12ms", "attempts": 1 }
  ]
}
```

`status` vaut `success`, `failed` ou `timeout`.

# CHIMERA — Mejoras pendientes

Lista priorizada de mejoras significativas. Seguridad/auth queda en baja prioridad porque CHIMERA opera en red cerrada.

## Alta prioridad (fiabilidad)

- [ ] Errores de I/O ignorados en el scheduler de envío UDP/TCP (`pkg/plate/scheduler.go:86`). Comprobar y registrar; pausar el ciclo si los fallos persisten.
- [ ] Fugas de goroutines en lectores de `EventCh` (`pkg/control/control.go:79`, `pkg/control/tui.go:91`) y handlers TCP del scheduler. Atar al `ctx` / `Delete()` con `WaitGroup`.
- [ ] Race en el broadcaster al cerrar el canal mientras otras goroutines escriben (`pkg/control/control.go:163`). Sincronizar con `WaitGroup` o flag atómico antes del `close`.
- [ ] Apagado limpio del modo TUI (`main.go:82`): bloquea en `p.Run()` y no propaga señales a las plates. Encadenar shutdown con `ctx`.
- [ ] Timeouts en accept/read TCP/UDP. Conexiones colgadas bloquean indefinidamente.
- [ ] Validación de comandos TUI (`set`): comprobar enum/rangos antes de aplicar y recoger panics en el handler.
- [ ] `StartControlDaemon` falla en silencio dentro de goroutine si el puerto está ocupado (`main.go:86`). Propagar el error al `main`.

## Media prioridad (calidad / mantenibilidad)

- [ ] Duplicación TUI local vs remoto (`pkg/control/tui.go` y `pkg/control/remote.go`, ~650 líneas casi idénticas). Extraer completer, dispatch y sync de boards a un módulo común.
- [ ] Estado mutable sin locks en `TUIServer` (`refreshBoardNames` vs ejecuciones concurrentes). Añadir `sync.RWMutex`.
- [ ] Errores ADJ silenciosos (`pkg/adj/adj.go:108` usa `println`, `pkg/adj/boards.go:119` idem). Wrapear con `fmt.Errorf` y validar campos requeridos al cargar.
- [ ] Validación de config incompleta (`pkg/config/config.go`): rango de puerto, existencia de interfaz, formato CIDR.
- [ ] Lookup de measurements case-sensitive mientras los nombres de board sí se normalizan. Unificar.
- [ ] Atomic swap del clone ADJ asume un único escritor (`pkg/adj/git.go`). Usar lock de fichero o directorios versionados.

## Baja prioridad (DX / despliegue / red)

- [ ] Hardcoded `127.0.0.1` en daemon y cliente remoto — solo si llega el caso de control multi-máquina.
- [ ] Auth en puerto de control — solo si la red deja de ser cerrada.
- [ ] Embeber versión/commit de CHIMERA (`-ldflags -X main.Version=...`) y mostrarla en el banner junto al ADJ.
- [ ] Tests: empezar por `pkg/adj` (parse/fallback) y parser de comandos en `pkg/control`.
- [ ] CI (GitHub Actions): `go vet`, `staticcheck`, `go test` en cada push.
- [ ] Dockerfile + ejemplo de despliegue documentando `CAP_NET_ADMIN` y la config de red.
- [ ] Constantes nombradas para magic numbers (buffer 256, capacidad de canal 16). Documentar el porqué.
- [ ] Limpiar código muerto: `Board.LookUpMeasurements` se construye pero nunca se lee.
- [ ] Métricas/observabilidad: contadores de paquetes enviados, errores y latencias del scheduler.

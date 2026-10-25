# Coding policies — sdd
<!-- Generado por /coding-policies el 2026-10-25. Regenerar con /coding-policies; "Ajustes de este proyecto" se preserva. -->
<!-- coding-policies: generated=2026-10-25; baseline=clean-code@2026-10-08; stacks=go@2026-09-14,typescript@2026-09-14,react@2026-10-09 -->

## Clean Code y SOLID
<!-- coding-policies:baseline=clean-code; version=2026-10-08 -->
### Responsabilidad única y cohesión

- **MUST** Aplicar SRP haciendo que cada función, componente, clase o módulo tenga una sola razón principal para cambiar y responda a un actor o grupo estrechamente relacionado. Porqué: "hacer una sola cosa" sin nombrar quién pide el cambio confunde pasos de una misma responsabilidad con responsabilidades realmente distintas. Gate: `revisión de diseño` que pueda describir responsabilidad y actor en una oración
- **MUST** Mantener juntas las partes que cambian por la misma razón y separar las que cambian por razones distintas. Porqué: SRP busca alta cohesión y bajo acoplamiento, no la mayor cantidad posible de archivos o funciones. Gate: `revisión del diff` y del historial de cambios relacionados
- **MUST** Separar decisiones de dominio, presentación, persistencia e integración cuando tengan actores, ritmos de cambio o contratos distintos; una función de orquestación puede coordinarlas sin implementar todas sus políticas. Porqué: coordinar colaboradores es una responsabilidad válida, mientras absorber sus detalles crea una unidad con múltiples motivos de cambio. Gate: `tests de contratos` en las fronteras identificadas
- **SHOULD** Usar cambios históricos, vocabulario, setup de tests y dependencias como evidencia de responsabilidades mezcladas antes de extraer. Porqué: una división fundada en señales reales conserva mejor la cohesión que una basada únicamente en la forma actual del código. Gate: `git log -- <archivo>` más revisión de imports y fixtures

### Tamaño como señal, no como objetivo

- **MUST** Tratar líneas, complejidad, anidamiento y cantidad de parámetros como detectores de humo, nunca como prueba suficiente de buen o mal diseño. Porqué: una unidad corta puede mezclar responsabilidades y una unidad larga puede ser una tabla declarativa perfectamente cohesionada. Gate: `revisión de responsabilidad y cohesión` junto con las métricas
- **MUST** Medir líneas lógicas con una convención estable que excluya líneas vacías y comentarios, y distinguir código ejecutable de datos o markup declarativo. Porqué: cambiar la forma de contar vuelve incomparables los umbrales y penaliza documentación o formato sin reducir complejidad. Gate: configuración versionada de `ESLint`, `detekt`, `golangci-lint` o herramienta equivalente ya adoptada
- **SHOULD** Considerar hasta 40 líneas lógicas como zona saludable habitual para una función, método o lógica ejecutable de un componente, sin exigir que toda unidad llegue a ese tamaño. Porqué: coincide con el punto de revisión de guías conservadoras y mantiene visible la intención sin promover microfunciones. Gate: `max-lines-per-function`, `LongMethod` o `funlen` configurado como advertencia
- **SHOULD** Tratar entre 41 y 60 líneas lógicas como señal de revisión, no como infracción automática. Porqué: ese rango coincide con defaults comunes de linters, pero todavía puede representar un algoritmo o flujo cohesivo cuya fragmentación empeore la lectura. Gate: `review del PR` asistido por la métrica del stack

### Funciones y métodos

- **MUST** Cuando una función, método o lógica ejecutable de componente tenga más de 60 líneas lógicas, evaluar su responsabilidad, complejidad y posibilidades de extracción; dividirla o documentar por qué mantener cohesión y localidad resulta más claro. Porqué: superar el rango usado por varios linters merece una decisión explícita, no un rechazo mecánico. Gate: `ESLint max-lines-per-function`, `detekt LongMethod` o `golangci-lint funlen` más justificación en el review
- **MUST** Mantener cada función en un nivel de abstracción reconocible y darle un nombre que exprese intención, sin mezclar orquestación de alto nivel con detalles incidentales extensos. Porqué: el lector debe entender primero qué ocurre y profundizar solo donde necesite saber cómo. Gate: `revisión de nombres y flujo` del caller y la función
- **SHOULD** Mantener la complejidad ciclomática en 10 o menos, tratar de 11 a 15 como señal y exigir revisión explícita por encima de 15, eligiendo guard clauses, tablas o extracción solo cuando aclaren el contrato. Porqué: más caminos independientes aumentan combinaciones y costo de prueba, pero reducir el número sin mejorar el modelo solo desplaza la complejidad. Gate: `ESLint complexity`, `detekt CyclomaticComplexMethod` o `golangci-lint cyclop`
- **SHOULD** Mantener el anidamiento en profundidad 3 o menos, tratar profundidad 4 como señal y revisar obligatoriamente cuando supere 4. Porqué: cada nivel agrega contexto simultáneo, aunque extraer bloques triviales puede ocultar el flujo en vez de simplificarlo. Gate: `ESLint max-depth`, `detekt NestedBlockDepth` o equivalente del stack
- **SHOULD** Preferir hasta 3 parámetros, tratar de 4 a 5 como señal y revisar más de 5, sin reemplazarlos mecánicamente por un objeto bolsa con campos no relacionados. Porqué: muchos parámetros suelen revelar responsabilidades o conceptos ausentes, pero agruparlos sin cohesión solo esconde la misma complejidad. Gate: `ESLint max-params`, `detekt LongParameterList` o revisión equivalente

### Componentes y UI

- **MUST** Dar a cada componente una responsabilidad de interfaz identificable y separar regiones que tengan estado, efectos, permisos, ciclos de carga o actores independientes. Porqué: un componente deja de ser cohesivo cuando cambios visuales sin relación pueden romper flujos y dependencias ajenos. Gate: `tests de componente` y revisión del ownership de estado y efectos
- **MUST** Aplicar los umbrales de función a la lógica ejecutable del componente, pero no usar la cantidad de JSX, Compose o markup declarativo como motivo único de extracción. Porqué: el markup puede ser largo y lineal sin agregar caminos ni responsabilidades, mientras pocas líneas con varios efectos pueden ser complejas. Gate: `revisión separada de lógica y markup` más linter del stack
- **SHOULD** Extraer un subcomponente cuando represente una región con nombre de dominio, semántica accesible, contrato de props, comportamiento testeable o reutilización real. Porqué: esas fronteras permiten razonar en aislamiento; cortar fragmentos arbitrarios solo agrega navegación y props. Gate: `tests por comportamiento` y revisión del árbol de componentes
- **MUST** Crear un custom Hook, presenter o componente auxiliar solo cuando encapsule una responsabilidad cohesiva; no mover líneas a un wrapper pasante para cumplir una métrica. Porqué: trasladar código sin reducir conocimiento compartido conserva la complejidad y suma una indirección. Gate: `review del contrato extraído` y de sus dependencias

### Archivos, clases y módulos

- **SHOULD** Considerar hasta 300 líneas lógicas como zona saludable habitual para archivos, clases o módulos escritos a mano. Porqué: coincide con un default extendido de lint y deja espacio para una unidad completa sin convertir cada concepto pequeño en un archivo. Gate: `ESLint max-lines`, `detekt LargeClass` o contador equivalente configurado como advertencia
- **SHOULD** Tratar entre 301 y 600 líneas lógicas como señal para revisar cohesión, superficie pública, dependencias y frecuencia de cambios, no como orden automática de dividir. Porqué: el tamaño intermedio puede indicar responsabilidades acumuladas o simplemente código declarativo que gana claridad al permanecer junto. Gate: `review del PR` con métrica y mapa de exports
- **MUST** Ante más de 600 líneas lógicas escritas a mano, tomar una decisión explícita de descomposición o justificar por qué una unidad cohesionada es más segura; código generado y grandes tablas declarativas se evalúan por su fuente y contrato. Porqué: a esa escala la navegación y el riesgo de mezclar cambios ameritan evidencia, pero una partición artificial también tiene costo. Gate: `revisión de arquitectura` registrada en el PR o excepción versionada del linter
- **MUST** Dividir un archivo o módulo cuando su API pública agrupe conceptos sin relación o cuando cambios de actores distintos lo modifiquen repetidamente, aunque todavía esté debajo del umbral de líneas. Porqué: SRP es una regla de cambio y cohesión, no una recompensa por archivos cortos. Gate: `git log -- <archivo>` más revisión de exports y consumidores

### SOLID sin ceremonia

- **MUST** Aplicar SRP como criterio de cohesión y razón de cambio, no como mandato de que cada unidad ejecute un único paso técnico. Porqué: una operación de negocio puede requerir varios pasos coordinados y seguir respondiendo a un solo actor. Gate: `revisión de responsabilidad` con actor, contrato e invariantes
- **SHOULD** Aplicar OCP creando puntos de extensión solo cuando exista variación observada o exigida por el contrato. Porqué: diseñar para todas las variantes imaginables produce abstracciones especulativas más difíciles de cambiar que el código directo. Gate: `tests de variantes` existentes o requisito que justifique la extensión
- **SHOULD** Aplicar LSP exigiendo que toda implementación sustituible preserve precondiciones, resultados, invariantes, errores y efectos observables del contrato. Porqué: compartir una interfaz o herencia no garantiza que los consumidores puedan reemplazar una implementación sin sorpresas. Gate: `contract tests` ejecutados contra cada implementación
- **SHOULD** Aplicar ISP definiendo interfaces desde las necesidades de sus consumidores y separándolas cuando obliguen a depender de operaciones ajenas. Porqué: contratos amplios aumentan acoplamiento, pero interfaces de un solo método sin consumidores distintos pueden ser ceremonia. Gate: `revisión de consumidores` y fakes de tests
- **SHOULD** Aplicar DIP haciendo que la política de alto nivel dependa de contratos estables e inyectando I/O, tiempo, azar, red y persistencia en sus fronteras volátiles. Porqué: aislar detalles cambiantes mejora pruebas y reemplazo sin invertir cada dependencia interna. Gate: `tests deterministas` con adapters reales cubiertos en integración
- **SHOULD** Evitar una interfaz, factory o capa por cada clase cuando no exista sustitución, frontera volátil ni consumidor que la necesite. Porqué: SOLID reduce acoplamiento útil; aplicado como plantilla multiplica archivos y saltos sin proteger ningún cambio real. Gate: `review de abstracciones` que nombre al menos una variación o frontera concreta

### Extracción sin sobre-split

- **MUST** Nunca extraer una unidad solo para reducir LOC; la extracción debe revelar intención, aislar una responsabilidad, reducir complejidad o establecer un contrato útil. Porqué: satisfacer una cifra sin mejorar comprensión convierte una función larga en una cadena de saltos. Gate: `review del antes y después` con el beneficio nombrado
- **MUST** Dar a toda unidad extraída un nombre específico del dominio o de la intención y entradas, salidas y efectos acotados. Porqué: nombres como `handlePart`, `processData` o `helper` desplazan líneas pero no explican responsabilidades. Gate: `revisión de nombres y firma` más tests del contrato
- **MUST** Preservar localidad cuando varios pasos comparten una invariante, orden o estado y se entienden mejor de forma secuencial. Porqué: dispersar un flujo cohesivo entre muchos archivos obliga al lector a reconstruir contexto sin reducir el conocimiento necesario. Gate: `review de navegación` desde el entry point hasta el resultado
- **SHOULD** Evitar wrappers pasantes, microarchivos y cadenas de funciones de una sola llamada salvo que marquen una frontera estable, mejoren el lenguaje o habiliten sustitución y tests. Porqué: cada indirección tiene costo cognitivo y debe comprar una separación verificable. Gate: `grafo de llamadas` o revisión de callers y contratos
- **SHOULD** Aplicar el test de borrado antes de crear o conservar un módulo: si al borrarlo la complejidad desaparece era un pasamanos, y si reaparece repetida en N llamadores se ganaba su lugar. Porqué: un módulo profundo concentra mucho comportamiento detrás de una interfaz chica, mientras que uno superficial solo suma un salto. Gate: `review del PR` que describa qué complejidad reaparece y en cuántos llamadores
- **SHOULD** No introducir un puerto o seam hasta tener dos adapters justificados, normalmente producción y test: un solo adapter es un seam hipotético y dos son un seam real. Porqué: una interfaz con una única implementación solo agrega indirección sin variación que proteger. Gate: `revisión de abstracciones` que nombre ambos adapters
- **SHOULD** Asumir que la interfaz es la superficie de test: si para testear hay que entrar por detrás de la interfaz, el módulo probablemente tiene la forma equivocada. Porqué: tests acoplados a internos se rompen al refactorizar sin que cambie el comportamiento observable. Gate: `tests por la interfaz pública` que sobrevivan a un refactor interno

### Smells como juicio

- **MUST** Usar los smells como juicio y nunca como regla dura: nombrarlos como "posible envidia de datos" y citar el fragmento; si una regla documentada del repo avala el patrón señalado, esa regla gana sobre el smell. Porqué: son heurísticas de Fowler que señalan dónde mirar, no violaciones demostrables. Gate: `review del PR` con el smell nombrado como hallazgo de juicio
- **MUST** No repetir como hallazgo lo que el formatter, el linter o el typechecker ya **falla**; las advertencias que la baseline deja como señal de revisión sí se evalúan. Porqué: duplicar una señal automática ensucia el review sin aportar criterio. Gate: `review del PR` sin hallazgos que dupliquen la salida de formatter, linter y typechecker
- **SHOULD** Revisar duplicación y datos sin tipo: código duplicado (la misma lógica, por la misma razón de cambio, en más de un lugar → extraer la forma común y llamarla desde ambos), grupos de datos (los mismos campos o parámetros viajan juntos → agruparlos en un tipo) y obsesión por primitivos (un primitivo o string hace de concepto de dominio → darle un tipo pequeño propio). Porqué: la repetición de forma indica un concepto que todavía no tiene nombre. Gate: `review del diff` buscando formas repetidas
- **SHOULD** Revisar acoplamiento entre objetos: envidia de datos (un método usa más datos ajenos que propios → moverlo hacia los datos), cadenas de mensajes (`a.b().c().d()` que el llamador no debería conocer → ocultar el recorrido tras un método del primer objeto) e intermediario (una clase cuya interfaz mayormente solo delega, sin marcar frontera, mejorar el lenguaje ni habilitar sustitución → eliminarla y llamar al destino real; ocultar el recorrido de una cadena de mensajes puede crear justo ese pasamanos, y el juicio decide cuál pesa más). Porqué: cada una hace que un cambio local se propague a quien solo debía pedir un resultado. Gate: `review de callers y contratos`
- **SHOULD** Revisar cómo se propaga el cambio: switches repetidos (el mismo `switch` o cascada de `if` sobre el mismo tipo en varios sitios → polimorfismo o un mapa compartido) y cambio divergente (un módulo se edita por razones no relacionadas → separarlo para que cambie por una sola). Porqué: un cambio de negocio no debería exigir editar sitios dispersos ni un archivo debería ser tocado por actores distintos. Gate: `git log -- <archivo>` más revisión del diff

### Adopción y excepciones

- **MUST** Respetar los límites y herramientas que el repositorio ya adoptó y no reconfigurarlos incidentalmente para cerrar una tarea. Porqué: los umbrales son una decisión de equipo y cambiarlos mezcla política global con comportamiento funcional. Gate: `git diff` de configuración de lint y CI
- **SHOULD** En proyectos sin métricas, introducir estos umbrales primero como warnings o revisión sobre código nuevo y endurecerlos solo con una baseline y acuerdo explícitos. Porqué: activar errores globales de una vez genera churn y refactors masivos sin relación con el cambio. Gate: `CI` con baseline o ratchet de archivos modificados
- **MUST** Acotar excepciones de tamaño a código generado, migraciones, fixtures o snapshots extensos, tablas declarativas y adapters mecánicos cuando dividirlos empeore trazabilidad; los tests no quedan exentos por categoría. Porqué: una excepción basada en naturaleza del contenido es revisable, mientras excluir carpetas enteras oculta lógica compleja. Gate: ignore específico en `ESLint`, `detekt`, `golangci-lint` o herramienta equivalente con motivo
- **MUST** Documentar toda supresión con alcance mínimo y razón concreta, y volver a evaluarla cuando cambie la responsabilidad de la unidad. Porqué: una excepción silenciosa se convierte en permiso permanente para acumular complejidad. Gate: `review de suppressions` y configuración versionada

### Lectura ampliada

- [Google C++ Style Guide: Write Short Functions](https://google.github.io/styleguide/cppguide.html#Write_Short_Functions) — recomienda funciones pequeñas y enfocadas, pero rechaza un límite duro y pide no dañar la estructura al dividir.
- [ESLint: max-lines-per-function](https://eslint.org/docs/latest/rules/max-lines-per-function) — regla configurable con default de 50 líneas por función.
- [ESLint: max-lines](https://eslint.org/docs/latest/rules/max-lines) — regla configurable con default de 300 líneas por archivo.
- [ESLint: complexity](https://eslint.org/docs/latest/rules/complexity) — medición configurable de complejidad ciclomática.
- [ESLint: max-depth](https://eslint.org/docs/latest/rules/max-depth) — profundidad máxima configurable, con default 4.
- [ESLint: max-params](https://eslint.org/docs/latest/rules/max-params) — cantidad máxima configurable, con default 3.
- [Detekt: Complexity Rule Set](https://detekt.dev/docs/rules/complexity/) — defaults para métodos largos, clases grandes, parámetros y complejidad en Kotlin.
- [golangci-lint: configuración de linters](https://golangci-lint.run/docs/linters/configuration/#funlen) — defaults de `funlen` y `cyclop` para Go.
- [React: Thinking in React](https://react.dev/learn/thinking-in-react) — separación de componentes por responsabilidad y crecimiento.
- [Martin Fowler: Function Length](https://martinfowler.com/bliki/FunctionLength.html) — extracción guiada por intención frente a implementación, no por una cifra aislada.
- [Refactoring: Extract Function](https://refactoring.com/catalog/extractFunction.html) — mecánica y motivación de una extracción con nombre significativo.
- [Robert C. Martin: The Single Responsibility Principle](https://blog.cleancoder.com/uncle-bob/2014/05/08/SingleReponsibilityPrinciple.html) — razón de cambio, actores, cohesión y separación de responsabilidades.
- [Robert C. Martin: SOLID Relevance](https://blog.cleancoder.com/uncle-bob/2020/10/18/Solid-Relevance.html) — alcance de los cinco principios y su aplicación a diseño y arquitectura.
- [Martin Fowler: Refactoring, cap. 3 "Bad Smells in Code"](https://martinfowler.com/books/refactoring.html) — catálogo de smells (duplicación, envidia de datos, grupos de datos, obsesión por primitivos, switches repetidos, cambio divergente, cadenas de mensajes, intermediario) como heurísticas de juicio.
- [John Ousterhout: A Philosophy of Software Design](https://web.stanford.edu/~ouster/cgi-bin/book.php) — módulos profundos: interfaces pequeñas con mucho comportamiento detrás, frente a módulos superficiales y pasamanos.

## Go
<!-- coding-policies:stack=go; version=2026-09-14 -->
### Layout de paquetes

- **MUST** `main` vive en `cmd/<binario>/main.go` (o en la raíz si el módulo tiene un solo binario) y solo hace wiring: lee la configuración, construye las dependencias y arranca. Porqué: lo que vale la pena testear vive en paquetes; un `main` gordo es lógica sin tests. Gate: —
- **MUST** Todo lo que no debe importarse desde afuera va bajo `internal/`; no crear `pkg/`. Porqué: el compilador hace cumplir el límite de API, y `pkg/` invita a exportar sin criterio. Gate: `go build ./...` (el compilador rechaza imports ajenos de `internal/`)
- **MUST** Nombrar cada paquete por lo que provee, en una palabra, minúscula y sin guiones; nunca `util`, `common`, `helpers`, `shared`, `base` ni `models`. Porqué: un paquete es una unidad de propósito; los cajones de sastre crecen sin límite y acoplan todo con todo. Gate: `find . -type d` (con nombres `util`, `common`, `helpers` o `shared` da vacío)
- **SHOULD** Separar kits reutilizables sin conocimiento del negocio (`internal/foundation/*`) de las features (`internal/business/*`); ninguna feature importa a otra, y los adapters entre features viven en `cmd/`. Porqué: la dependencia cruzada entre features es el primer paso al monolito enredado; el adapter en `cmd/` la hace visible y revisable. Gate: `golangci-lint` (depguard con las reglas foundation ∤ business y business ∤ business)

### Errores

- **MUST** Envolver con `fmt.Errorf("<paquete>: <operación>: %w", err)` al cruzar una frontera; nunca `%v` ni `err.Error()` para envolver. Porqué: `%w` conserva la cadena para `errors.Is` y `errors.As`, y el prefijo dice dónde falló sin stack trace. Gate: `golangci-lint` (errorlint)
- **MUST** Manejar todo error devuelto: se retorna, o se maneja y se sigue con el motivo escrito en el sitio; nunca `_ = f()` sin comentario. Porqué: un error ignorado es un bug diferido que aparece lejos de su causa. Gate: `golangci-lint` (errcheck)
- **MUST** Un error se maneja una sola vez: o se loggea o se retorna, nunca ambos; el log vive en el borde (middleware, `main`). Porqué: el doble manejo duplica logs y confunde quién es dueño del error. Gate: —
- **MUST** Sentinels `var ErrX = errors.New("...")` para condiciones que el caller decide con `errors.Is`; tipos de error propios solo cuando el caller necesita datos (`errors.As`); jamás comparar strings de error. Porqué: el matching por string se rompe con cualquier cambio de mensaje. Gate: `golangci-lint` (errorlint)
- **SHOULD** Mensajes de error en minúscula, sin puntuación final y accionables: dicen qué pasó y qué hacer. Porqué: se concatenan en cadenas `a: b: c`, y quien los lee tiene que poder actuar. Gate: `golangci-lint` (staticcheck: ST1005)
- **SHOULD** Mapear errores a códigos HTTP o exit codes en la frontera (handler o `main`) con una tabla explícita por sentinel y un fallback genérico que no exponga el error crudo. Porqué: el service no sabe de HTTP, y un 500 con el error interno filtra detalles al cliente. Gate: —
- **SHOULD** Modelar "no encontrado" que no es falla con `(valor, ok bool, err error)` o un sentinel `ErrNotFound`; nunca `nil, nil` sin documentar. Porqué: el caller necesita distinguir ausencia de fallo, y un 401 silenciado como "no existe" esconde un problema de credenciales. Gate: —
- **MUST** No usar `panic` para errores esperables fuera de `main`; `recover` solo en el borde de una goroutine que no puede tirar el proceso. Porqué: un panic en un paquete mata al proceso ajeno que lo importa. Gate: `golangci-lint` (forbidigo: `panic(` fuera de `main`)

### Naming

- **MUST** Receivers de una o dos letras, consistentes en todos los métodos del tipo (`s *service`, `r *repo`); nunca `this` ni `self`. Porqué: el receiver se lee cientos de veces y su nombre no aporta información. Gate: `golangci-lint` (revive: receiver-naming)
- **MUST** Acrónimos en mayúscula sostenida en identificadores: `ID`, `URL`, `HTTP`, `API`, `DNI` (`userID`, `PhotoURL`, nunca `userId`). Porqué: es la convención del ecosistema y lo que los linters esperan. Gate: `golangci-lint` (revive: var-naming)
- **MUST** Getters sin prefijo `Get`: `Slug()`, no `GetSlug()`; `Get*` se reserva para operaciones con I/O (`GetByID`). Porqué: el prefijo no agrega información y ensucia el API. Gate: —
- **MUST** Evitar el stutter paquete.Tipo: `registry.Client`, no `registry.RegistryClient`. Porqué: el nombre siempre se lee calificado por el paquete. Gate: `golangci-lint` (revive: exported)
- **SHOULD** Interfaces de un método con sufijo `-er` (`TagLister`, `Mailer`); los puertos principales con el sustantivo del rol (`Repository`, `Service`, `Driver`). Porqué: el nombre dice el rol sin abrir la definición. Gate: —
- **SHOULD** Variables cortas en scopes cortos (`i`, `ctx`, `dev`, `fs`) y nombres más largos cuanto más lejos se usan; constantes agrupadas por familia con prefijo común (`LabelProject`, `LabelService`). Porqué: la longitud del nombre es proporcional a la distancia entre declaración y uso. Gate: —

### Interfaces y tipos

- **MUST** Interfaces chicas (1 a 3 métodos) declaradas del lado del consumidor cuando cruzan paquetes, con su propio DTO de proyección; el puerto grande de un paquete (`Repository`, `Service`, `Driver`) vive junto a su implementación, documentado por grupos de métodos. Porqué: el consumidor depende solo de lo que usa, y el fake queda trivial. Gate: —
- **MUST** Constructores `New...` que reciben las dependencias por parámetro; sin `init()` con efectos ni estado global mutable. Porqué: las dependencias explícitas se testean y se reemplazan; los globals esconden el grafo. Gate: `golangci-lint` (gochecknoinits, gochecknoglobals)
- **SHOULD** Lo opcional se expresa con campos exportados nil-ables documentados (`// optional; nil disables`) o constructores alternativos explícitos (`NewTo`, `NewServiceWithClock`); sin functional options. Porqué: las options esconden qué combinaciones son válidas y agregan código; un campo `nil` documentado se lee en el struct. Gate: —
- **MUST** Receiver puntero para tipos con estado o mutables; receiver de valor solo para tipos-valor chicos e inmutables (`LocalDate`, `Device`); nunca mezclar en un mismo tipo. Porqué: mezclar produce copias inesperadas y métodos que no ven mutaciones. Gate: `golangci-lint` (recvcheck)
- **MUST** Mutex como campo nombrado `mu`, primero en el struct, nunca embebido ni exportado; el lock cubre exactamente los campos que protege. Porqué: embeberlo exporta `Lock` y `Unlock` en el API público del tipo. Gate: —

### Concurrencia

- **MUST** `ctx context.Context` es el primer parámetro de toda función que hace I/O; se propaga, no se guarda en structs, y se nombra `_` si no se usa. Porqué: la cancelación y los deadlines viajan por el ctx; guardarlo en un struct lo desacopla del request. Gate: `golangci-lint` (contextcheck, containedctx)
- **MUST** Toda goroutine tiene dueño y salida: `sync.WaitGroup` o `errgroup` en el mismo scope, o un `select` sobre `ctx.Done()` y un canal de stop; un proceso deliberadamente long-lived se documenta en el sitio. Porqué: una goroutine sin dueño es un leak y un panic invisible. Gate: —
- **MUST** Pasar las variables del loop como parámetros a la goroutine (`go func(i int, d Driver) {...}(i, d)`), y transportar valor y error por índice en un slice pre-dimensionado antes que por canal. Porqué: la captura del iterador se lee ambigua aunque Go 1.22 la haya corregido, y el slice por índice evita mutex y conserva el orden. Gate: `go vet` (loopclosure)
- **MUST** Timeouts explícitos en cada frontera de I/O: `context.WithTimeout` por operación de repositorio o cliente, `http.Client{Timeout: ...}`, `ReadHeaderTimeout` en servidores, y `cancel()` siempre invocado. Porqué: sin timeout, un socket colgado bloquea la request y drena el pool. Gate: `go vet` (lostcancel) y `golangci-lint` (noctx, bodyclose)
- **MUST** Shutdown ordenado en `main`: `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)` y `srv.Shutdown` con un timeout acotado. Porqué: perder requests en vuelo en cada deploy es un bug reproducible. Gate: —
- **SHOULD** Reloj inyectable (`now func() time.Time`, default `time.Now`) en servicios que dependen del tiempo, y `time.Now().UTC()` al persistir. Porqué: sin seam de reloj los tests de fechas son frágiles; sin UTC, la hora depende del host. Gate: —

### Testing

- **MUST** Tests table-driven con `t.Run(tc.name, ...)`, nombres de caso en prosa que describan el escenario, y `TestFunc_Escenario` como nombre del test. Porqué: el nombre es la spec; un fallo se entiende sin abrir el código. Gate: —
- **MUST** Aserciones con la stdlib (`t.Fatalf("x = %q, want %q", got, want)`) o `testify/require`; nunca `assert` que deje seguir un test ya roto. Porqué: un assert que no frena produce cascadas de fallos falsos. Gate: `golangci-lint` (testifylint)
- **MUST** Dobles escritos a mano que implementan la interfaz, con contadores de llamadas y errores inyectables, extendidos por embedding para cada escenario; sin gomock ni mockery. Porqué: el fake se lee, se adapta en el mismo archivo y no regenera código. Gate: —
- **MUST** Aislar con recursos reales baratos antes que con mocks: `httptest`, `t.TempDir()` con SQLite, el binario real; mocks solo en límites de sistema. Porqué: el test que habla con lo real detecta el bug que el mock oculta. Gate: —
- **MUST** `t.Helper()` en todo helper compartido y `t.Cleanup()` para liberar recursos. Porqué: sin `Helper` el fallo apunta al helper y no al test; `Cleanup` corre aunque el test haga `Fatal`. Gate: `golangci-lint` (thelper)
- **SHOULD** Extraer el núcleo puro de toda función que toca el mundo (dispositivo, red, base) a una función sin I/O y testearla directo; `export_test.go` como único seam white-box. Porqué: el test unitario corre en milisegundos y sin fixtures, y el resto se cubre con integración. Gate: —

### Dependencias y configuración

- **MUST** Preferir la stdlib; cada dependencia directa nueva se justifica en el PR (qué no resuelve `net/http`, `flag`, `log/slog`, `database/sql`) y se revisa su árbol transitivo. Porqué: cada dependencia es superficie de ataque, tiempo de build y deuda de upgrade. Gate: `go mod why` (revisar el diff de `go.mod` en el PR)
- **MUST** `go.mod` limpio: `go mod tidy` sin diff, sin `replace` a rutas locales, y la directiva `go` alineada con la versión que corre CI. Porqué: un `go.sum` con drift rompe builds reproducibles. Gate: `go mod tidy` (`git diff --exit-code go.mod go.sum`)
- **MUST** Configuración por variables de entorno leídas en `main` o en un paquete `config` con `Validate()` fail-closed, y pasadas como valores; nunca `os.Getenv` disperso en la lógica. Porqué: la configuración dispersa no se testea ni se documenta. Gate: `golangci-lint` (forbidigo: `os.Getenv` fuera de `main` y `config`)
- **MUST** `gofmt` y `go vet` limpios en CI; imports en tres grupos: stdlib, terceros, módulo propio. Porqué: es el mínimo que todo el ecosistema asume y lo que evita diffs de formato. Gate: `gofmt -l .` (vacío) y `go vet ./...`
- **SHOULD** `golangci-lint` con lista explícita de linters (`default: none` más `errcheck`, `govet`, `staticcheck`, `gosec`, `errorlint`, `revive`) y versión pineada en CI. Porqué: el set default cambia entre versiones y rompe CI sin tocar código. Gate: `golangci-lint run`

### Lectura ampliada

- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md) — la guía base de estilo: errores, interfaces, concurrencia, tests.
- [Effective Go](https://go.dev/doc/effective_go) — idioms del lenguaje: nombres, getters, interfaces, embedding.
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) — checklist de review: receivers, acrónimos, contexts, error strings.
- [Package Oriented Design](https://www.ardanlabs.com/blog/2017/02/package-oriented-design.html) — layout `cmd/` e `internal/`, kits y business; el porqué de no tener `util`.

## TypeScript
<!-- coding-policies:stack=typescript; version=2026-09-14 -->
### Alcance y configuración

- **MUST** Antes de editar, identificar la versión de TypeScript, el gestor y lockfile, los `tsconfig` efectivos, la configuración de lint, los scripts de CI, el runtime de producción y el pipeline de compilación. Porqué: la sintaxis válida y el significado de módulos, paths y emisión dependen del entorno real, no de la versión más nueva de la documentación. Gate: —
- **MUST** Respetar los scripts y convenciones existentes; una corrección acotada no migra módulos, dependencias, formatter, linter ni arquitectura como efecto incidental. Porqué: mezclar una migración con un cambio funcional amplía el riesgo y vuelve ambiguo cualquier fallo. Gate: `git diff -- package.json '*lock*' 'tsconfig*.json'`
- **MUST** Consultar documentación compatible con las versiones instaladas antes de adoptar una API, opción de compilador o preset de lint. Porqué: las páginas vivas pueden documentar capacidades o defaults que el proyecto todavía no tiene. Gate: —
- **MUST** Conservar las comprobaciones existentes y corregir sus errores; no desactivar `strict`, una regla type-aware ni una opción adicional para cerrar la tarea. Porqué: bajar el nivel de verificación convierte el error visible en riesgo silencioso para todo el proyecto. Gate: `tsc --showConfig -p <tsconfig>`
- **SHOULD** En proyectos nuevos, declarar `strict: true` aunque la versión instalada ya lo active por defecto. Porqué: explicita la intención y evita que el contrato dependa de un default histórico del compilador. Gate: `tsc --showConfig -p <tsconfig>`
- **SHOULD** En proyectos nuevos, evaluar y decidir explícitamente `noUncheckedIndexedAccess`, `exactOptionalPropertyTypes`, `noImplicitOverride` y `noFallthroughCasesInSwitch`; en proyectos existentes, incorporarlos solo mediante un cambio acotado. Porqué: cubren ausencias y contratos que `strict` no modela por completo, pero habilitarlos incidentalmente puede exigir una migración amplia. Gate: `tsc --showConfig -p <tsconfig>`
- **MUST** Ejecutar el typecheck con el proyecto correcto; no pasar archivos sueltos a `tsc` cuando hay un `tsconfig`, y respetar el flujo de project references si existe. Porqué: invocar el compilador fuera del proyecto puede ignorar opciones y producir una señal falsa. Gate: `tsc --noEmit -p <tsconfig>` o el script de typecheck del repo

### Tipos e inferencia

- **SHOULD** Usar inferencia para valores locales obvios y explicitar contratos exportados o fronteras cuando mejore la revisión. Porqué: repetir tipos locales agrega ruido, mientras una API explícita comunica qué estabilidad promete. Gate: `tsc --noEmit -p <tsconfig>`
- **MUST** No usar `as T`, `!` ni una anotación para inventar garantías sobre JSON, red o valores posiblemente ausentes; demostrar el narrowing o validar en runtime. Porqué: las assertions desaparecen al ejecutar y pueden ocultar exactamente el caso que el tipo advertía. Gate: `typescript-eslint` (`no-non-null-assertion` y reglas `no-unsafe-*`)
- **MUST** Representar valores desconocidos con `unknown` y reducirlos antes de operar; no abrir el contrato con `any`. Porqué: `unknown` conserva la obligación de comprobar el valor, mientras `any` propaga operaciones sin verificar. Gate: `typescript-eslint` (`no-explicit-any`, `no-unsafe-assignment`, `no-unsafe-argument`, `no-unsafe-call`, `no-unsafe-member-access`, `no-unsafe-return`)
- **MUST** Si una integración obliga a usar `any`, aislarlo en la frontera mínima y explicar por qué es inevitable; nunca encubrirlo con `as unknown as T`. Porqué: un escape localizado limita la pérdida de seguridad y deja visible la deuda. Gate: `typescript-eslint` (`no-explicit-any`) y revisión de `as unknown as`
- **SHOULD** Elegir `type` o `interface` por sus necesidades de composición y declaration merging, sin conversiones masivas de estilo. Porqué: ambas construcciones son válidas y una preferencia cosmética no justifica churn ni riesgo. Gate: —

### Modelado y narrowing

- **MUST** Modelar estados con datos distintos como uniones discriminadas en vez de grupos de propiedades opcionales independientes. Porqué: el tipo debe permitir solo combinaciones válidas y llevar los datos propios de cada variante. Gate: `tsc --noEmit -p <tsconfig>`
- **MUST** Reducir tipos mediante comprobaciones reales; todo type predicate manual debe implementar y testear la validación que declara. Porqué: la firma de un predicate puede mentirle al compilador aunque su cuerpo acepte valores inválidos. Gate: `tsc --noEmit -p <tsconfig>` y tests del predicate
- **MUST** Tratar las variantes cerradas de forma exhaustiva con `never` o lint; no esconder casos nuevos detrás de un `default` genérico salvo fallback defensivo exigido por datos de runtime. Porqué: agregar una variante debe señalar cada lugar que necesita una decisión. Gate: `typescript-eslint` (`switch-exhaustiveness-check`)
- **SHOULD** Usar `satisfies` para comprobar mapas de configuración y registros conservando sus literales específicos cuando esa inferencia sea útil. Porqué: valida el contrato sin ensanchar innecesariamente el valor como puede hacerlo una anotación. Gate: `tsc --noEmit -p <tsconfig>`

### Genéricos y APIs

- **MUST** Agregar un genérico solo cuando exprese una relación útil entre dos o más posiciones del contrato; evitar parámetros que aparecen una sola vez o no restringen nada. Porqué: un genérico sin relación agrega abstracción sin aportar información al caller. Gate: `typescript-eslint` (`no-unnecessary-type-parameters`)
- **SHOULD** Preferir una unión a overloads cuando ambas expresen el mismo contrato, y evitar restricciones genéricas más fuertes de lo necesario. Porqué: la firma más simple tiene menos casos divergentes entre declaración e implementación. Gate: `typescript-eslint` (`unified-signatures`)
- **MUST** Conservar en el tipo la relación real entre entrada y salida cuando el API la promete; no devolver un union amplio que obligue al caller a reconstruir una correlación conocida. Porqué: perder esa relación desplaza assertions y branches inseguros a todos los consumidores. Gate: `tsc --noEmit -p <tsconfig>`

### Límites y validación

- **MUST** Validar entradas externas relevantes al cruzar el límite del sistema y usar desde allí el valor validado. Porqué: una interfaz TypeScript no comprueba JSON, variables de entorno, storage ni respuestas de red en runtime. Gate: `runner de tests del repo` con entradas válidas e inválidas
- **MUST** Reutilizar el validador ya adoptado por el proyecto; la necesidad de validar no autoriza instalar Zod u otra dependencia por sí sola. Porqué: duplicar mecanismos aumenta bundle, mantenimiento y semánticas de error. Gate: `git diff -- package.json '*lock*'`
- **SHOULD** Derivar tipos desde el esquema cuando la librería lo soporte y distinguir input de output si hay coerciones o transformaciones. Porqué: un contrato duplicado puede driftear y una transformación hace que el tipo recibido no sea el tipo producido. Gate: `tsc --noEmit -p <tsconfig>`
- **MUST** Elegir deliberadamente qué hacer con campos desconocidos y revisar la semántica de toda coerción; no asumir que strings como `"false"` se convierten según intención humana. Porqué: strip, rechazo, passthrough y coerción cambian el contrato y pueden aceptar datos inesperados. Gate: `runner de tests del repo` con campos extra y coerciones

### Promesas y concurrencia

- **MUST** Toda promesa se espera, se retorna al caller o maneja explícitamente su rechazo. Porqué: una promesa flotante pierde errores y puede dejar efectos incompletos fuera del flujo observable. Gate: `typescript-eslint` (`no-floating-promises`)
- **MUST** No considerar `void tarea()` como manejo de errores; si se desprende trabajo, definir dónde se observa y trata el rechazo. Porqué: `void` puede silenciar al linter pero no cambia la semántica de la promesa. Gate: `typescript-eslint` (`no-floating-promises`) con `ignoreVoid: false`, o revisión equivalente
- **MUST** No pasar callbacks `async` a APIs que ignoran su promesa, como `forEach`; usar un loop secuencial o una agregación de promesas según la semántica. Porqué: el caller no espera esos callbacks y sus fallos quedan desacoplados. Gate: `typescript-eslint` (`no-misused-promises`)
- **MUST** Elegir conscientemente entre ejecución secuencial y concurrente, preservar el orden cuando importe y limitar fan-out contra servicios externos. Porqué: `Promise.all` sin criterio puede romper dependencias, saturar recursos o cambiar el comportamiento observable. Gate: `runner de tests del repo` con orden, fallo parcial y límite de concurrencia

### Módulos y runtime

- **MUST** Alinear `module`, `moduleResolution`, `package.json`, extensiones e imports con quien ejecuta o bundlea el JavaScript final. Porqué: que el dev server resuelva un import no demuestra que Node, el bundler o los consumidores de una librería puedan hacerlo. Gate: `tsc` más build o ejecución del artefacto de producción
- **SHOULD** Usar `import type` para dependencias que solo existen en el sistema de tipos, respetando el pipeline y `verbatimModuleSyntax` del proyecto. Porqué: hace explícita la frontera de runtime y evita emisiones o imports ambiguos. Gate: `typescript-eslint` (`consistent-type-imports`) o `tsc --noEmit -p <tsconfig>`
- **SHOULD** Separar `tsconfig` cuando servidor, navegador, workers y tests necesiten globals o emisión distintos. Porqué: una configuración única demasiado amplia permite APIs que no existen en alguno de los runtimes. Gate: `tsc -b` o scripts de typecheck por entorno
- **MUST** Si Node ejecuta TypeScript nativo, mantener un typecheck independiente y no asumir que el type stripping aplica aliases de `paths` ni transformaciones. Porqué: la ejecución nativa borra tipos pero no usa el `tsconfig` para comprobarlos o resolverlos. Gate: `tsc --noEmit -p <tsconfig>` y ejecución con la versión real de Node

### Linting y supresiones

- **SHOULD** Si el repo usa ESLint, partir de `recommendedTypeChecked` y configurar obtención de tipos con `projectService: true` cuando las versiones instaladas lo soporten. Porqué: las reglas type-aware detectan propagación insegura que el lint sintáctico no ve. Gate: `eslint .`
- **MUST** Adoptar `strictTypeChecked` solo deliberadamente y no habilitar el preset `all` automáticamente. Porqué: los presets opinados pueden cambiar fuera de una major y `all` introduce reglas sin una decisión de equipo. Gate: revisión de la configuración de `typescript-eslint` y `eslint .`
- **MUST** Corregir el problema antes de suprimirlo; una excepción usa `@ts-expect-error` con motivo concreto y alcance mínimo, nunca desactiva chequeos globales para cerrar la tarea. Porqué: `@ts-expect-error` falla cuando la excepción deja de ser necesaria y la explicación conserva el contexto. Gate: `typescript-eslint` (`ban-ts-comment`)
- **MUST** Verificar la matriz soportada de TypeScript, ESLint, parser y Node antes de actualizar cualquiera de ellos. Porqué: que el gestor resuelva versiones no significa que la combinación tenga soporte oficial. Gate: lockfile más documentación de compatibilidad de `typescript-eslint`

### Testing y verificación

- **MUST** Ejecutar los scripts pertinentes del repo para typecheck, lint, tests y build; si no hay typecheck y aplica, usar el compilador local con `--noEmit` y el `tsconfig` correcto. Porqué: cada gate cubre defectos distintos y una herramienta global puede no coincidir con el proyecto. Gate: scripts de CI del repo y `tsc --noEmit -p <tsconfig>`
- **MUST** Probar comportamiento observable y casos de error relevantes, especialmente validación, narrowing manual, cancelación y concurrencia cuando cambien. Porqué: que el código compile no demuestra que el contrato de runtime se cumpla. Gate: `runner de tests del repo`
- **MUST** Informar qué verificaciones se ejecutaron y cuáles no; no presentar un build sin typecheck ni una ejecución con type stripping como evidencia de tipos. Porqué: una señal mal etiquetada crea una confianza que la herramienta no produjo. Gate: `reporte final` con comando y resultado

### Rendimiento del tipado

- **SHOULD** Nombrar cálculos de tipos complejos reutilizados y preferir composición legible; evaluar `interface extends` frente a grandes intersecciones de objetos. Porqué: tipos anónimos e intersecciones profundas pueden repetir trabajo y empeorar diagnósticos. Gate: `tsc --noEmit -p <tsconfig>`
- **MUST** Medir antes de refactorizar por rendimiento del sistema de tipos. Porqué: sin diagnóstico se puede añadir complejidad sin atacar el cuello de botella real. Gate: `tsc --extendedDiagnostics -p <tsconfig>` o trace del compilador
- **MUST** No activar `skipLibCheck` automáticamente para ocultar conflictos entre dependencias; investigar el origen y documentar el compromiso si se adopta. Porqué: acelera el chequeo omitiendo errores en declaraciones y reduce la cobertura del compilador. Gate: `tsc --showConfig -p <tsconfig>` y typecheck con la decisión documentada

### Lectura ampliada

- [TypeScript: strict](https://www.typescriptlang.org/tsconfig/strict.html) — familia base de comprobaciones estrictas.
- [TypeScript: noUncheckedIndexedAccess](https://www.typescriptlang.org/tsconfig/noUncheckedIndexedAccess.html) — ausencia posible en accesos indexados.
- [TypeScript: exactOptionalPropertyTypes](https://www.typescriptlang.org/tsconfig/exactOptionalPropertyTypes.html) — diferencia entre propiedad ausente y `undefined` explícito.
- [TypeScript Handbook: Everyday Types](https://www.typescriptlang.org/docs/handbook/2/everyday-types.html) — inferencia, annotations, unions y assertions.
- [TypeScript Handbook: Narrowing](https://www.typescriptlang.org/docs/handbook/2/narrowing.html) — guards, predicates, uniones discriminadas y exhaustividad.
- [TypeScript Handbook: More on Functions](https://www.typescriptlang.org/docs/handbook/2/functions.html) — genéricos y overloads simples.
- [TypeScript: `satisfies`](https://www.typescriptlang.org/docs/handbook/release-notes/typescript-4-9.html#the-satisfies-operator) — comprobación sin perder inferencia específica.
- [typescript-eslint: typed linting](https://typescript-eslint.io/getting-started/typed-linting/) — configuración de reglas con información de tipos.
- [typescript-eslint: shared configs](https://typescript-eslint.io/users/configs/) — alcance y estabilidad de los presets.
- [typescript-eslint: avoiding `any`](https://typescript-eslint.io/blog/avoiding-anys/) — reglas que limitan `any` explícito y propagado.
- [typescript-eslint: no-floating-promises](https://typescript-eslint.io/rules/no-floating-promises/) — promesas cuyo resultado no se observa.
- [typescript-eslint: no-misused-promises](https://typescript-eslint.io/rules/no-misused-promises/) — promesas en posiciones que esperan valores síncronos.
- [typescript-eslint: switch-exhaustiveness-check](https://typescript-eslint.io/rules/switch-exhaustiveness-check/) — exhaustividad type-aware.
- [TypeScript: choosing compiler options for modules](https://www.typescriptlang.org/docs/handbook/modules/guides/choosing-compiler-options.html) — configuración según Node, bundler o librería.
- [TypeScript: verbatimModuleSyntax](https://www.typescriptlang.org/tsconfig/verbatimModuleSyntax.html) — relación explícita entre imports y emisión.
- [Node.js: Modules TypeScript](https://nodejs.org/api/typescript.html) — capacidades y límites de ejecutar TypeScript nativo.
- [TypeScript: erasableSyntaxOnly](https://www.typescriptlang.org/tsconfig/erasableSyntaxOnly.html) — sintaxis compatible con borrado de tipos.
- [Zod: Basic usage](https://zod.dev/basics) — validación de entradas y derivación de tipos cuando el repo usa Zod.
- [TypeScript Performance](https://github.com/microsoft/TypeScript/wiki/Performance) — diagnóstico y diseño de tipos con mejor desempeño.
- [TypeScript: skipLibCheck](https://www.typescriptlang.org/tsconfig/skipLibCheck.html) — alcance y pérdida de cobertura de la opción.
- [TypeScript 6.0 release notes](https://www.typescriptlang.org/docs/handbook/release-notes/typescript-6-0.html) — cambios de defaults y deprecaciones que requieren revisar versiones.

## React
<!-- coding-policies:stack=react; version=2026-10-09 -->
### Pureza y Hooks

- **MUST** Mantener puro el render: la misma combinación de props, estado y contexto produce el mismo JSX y no ejecuta efectos secundarios. Porqué: React puede renderizar, pausar o repetir trabajo antes de confirmar una actualización. Gate: `eslint-plugin-react-hooks` y tests de comportamiento
- **MUST** No mutar props, estado ni valores recibidos por Hooks. Porqué: React trata esos valores como snapshots y la mutación rompe detección de cambios y razonamiento entre renders. Gate: `reglas de inmutabilidad del linter o revisión del diff`
- **MUST** Usar componentes mediante JSX en vez de invocarlos como funciones ordinarias. Porqué: React necesita controlar su identidad, Hooks y lifecycle dentro del árbol. Gate: `eslint-plugin-react-hooks` más tests de render
- **MUST** Llamar Hooks ordinarios en el nivel superior de componentes o custom Hooks; la API `use` puede admitir condiciones o loops, pero sigue dentro de un componente o Hook y fuera de `try`/`catch`. Porqué: el orden estable asocia cada llamada con su estado y las excepciones de `use` no se extienden a `useState` o `useEffect`. Gate: `eslint-plugin-react-hooks` (`rules-of-hooks`)
- **MUST** Conservar las reglas recomendadas de `eslint-plugin-react-hooks` compatibles con la versión instalada y no desactivarlas para cerrar la tarea. Porqué: codifican invariantes de React que el typecheck no observa. Gate: `eslint .`

### Estado y reducers

- **MUST** Mantener el estado mínimo y evitar guardar valores que pueden calcularse desde props y estado durante render. Porqué: dos fuentes de verdad se desincronizan y agregan Effects innecesarios. Gate: revisión de `useState` y `useReducer`
- **SHOULD** Guardar la identidad seleccionada cuando el objeto completo ya vive en otra colección. Porqué: conservar una copia del objeto permite que quede obsoleta respecto de la fuente. Gate: `tests de actualización de la colección`
- **MUST** No copiar props a estado salvo que representen deliberadamente un valor inicial y el nombre haga explícita esa semántica. Porqué: el estado copiado deja de seguir cambios del parent sin que el contrato lo diga. Gate: `tests de cambio de props`
- **MUST** Modelar variantes de estado de modo que no permitan combinaciones contradictorias, usando discriminantes cuando cada estado tenga datos distintos. Porqué: booleans y campos opcionales independientes permiten loading, éxito y error al mismo tiempo. Gate: `typecheck y tests de transiciones`
- **SHOULD** Introducir un reducer cuando reúna transiciones complejas o dispersas; sus acciones describen eventos y su implementación permanece pura, mientras casos simples siguen con `useState`. Porqué: el reducer sirve para hacer comprensible un flujo, no como ceremonia universal. Gate: `tests de transiciones del reducer`

### Effects

- **MUST** Usar Effects para sincronizar con sistemas externos, no como mecanismo general de flujo interno. Porqué: un Effect agrega un ciclo posterior al render y puede ejecutarse más veces de las imaginadas. Gate: revisión de cada `useEffect`
- **MUST** Calcular datos derivados durante render y ejecutar interacciones desde sus event handlers, sin Effects que copien listas filtradas, encadenen estado o disparen una compra. Porqué: el lugar causal directo evita renders extra y acciones que ocurren por haber pintado UI. Gate: `tests de interacción y revisión de Effects`
- **SHOULD** Para datos remotos, preferir el mecanismo del framework o la solución existente que gestione cache, deduplicación y concurrencia. Porqué: un fetch manual en un Effect suele reimplementar lifecycle y carreras de red. Gate: `revisión de la capa de datos del repo`
- **MUST** Limpiar simétricamente suscripciones, conexiones y timers creados por un Effect; ante requests manuales, cancelar o ignorar respuestas obsoletas. Porqué: setup → cleanup → setup debe funcionar sin leaks ni resultados fuera de orden. Gate: `tests de cleanup y respuestas tardías`
- **MUST** Declarar como dependencias todos los valores reactivos leídos y no suprimir `exhaustive-deps` para imponer un array vacío. Porqué: una dependencia falsa deja closures viejos y esconde un diseño de sincronización incorrecto. Gate: `eslint-plugin-react-hooks` (`exhaustive-deps`)
- **SHOULD** Si la versión lo soporta, usar `useEffectEvent` solo para lógica no reactiva disparada desde Effects; no usarlo para ocultar dependencias, pasarlo a otros componentes ni llamarlo desde handlers ordinarios. Porqué: separa lecturas recientes de la sincronización sin convertirlo en un escape del modelo reactivo. Gate: `eslint-plugin-react-hooks` y tests del Effect
- **MUST** Corregir los problemas que revela Strict Mode antes de plantear desactivarlo. Porqué: la repetición de setup y cleanup expone Effects no idempotentes que también fallan en navegación real. Gate: tests bajo `StrictMode` o entorno de desarrollo equivalente

### Identidad y componentes

- **MUST** Usar `key` cuando una nueva identidad del dominio deba reiniciar el estado de un subárbol. Porqué: React preserva estado por tipo y posición; una key comunica que ya no es la misma entidad. Gate: `test de cambio de identidad`
- **MUST** No declarar componentes dentro de otros componentes si deben conservar identidad entre renders. Porqué: cada render crea un tipo distinto y React desmonta su estado. Gate: `lint de componentes anidados o revisión del árbol`
- **SHOULD** Mantener el estado en el ancestro común más cercano que realmente coordina a sus consumidores, sin elevarlo por defecto a un store global. Porqué: ownership local reduce acoplamiento y renders ajenos. Gate: `profiler y revisión del flujo de datos`

### Memoización

- **MUST** Comprobar si React Compiler está habilitado antes de diseñar optimizaciones manuales nuevas. Porqué: una versión moderna de React no demuestra que el compilador esté configurado para ese código. Gate: configuración y diagnóstico de `React Compiler`
- **SHOULD** Agregar `useMemo`, `useCallback` o `memo` solo por una necesidad concreta y medible. Porqué: también agregan dependencias y complejidad, y pueden costar más que el cálculo evitado. Gate: `profiler antes y después`
- **MUST** No retirar memoización existente en bloque sin comprobar comportamiento y rendimiento. Porqué: puede formar parte de un contrato de identidad o proteger un hotspot no evidente. Gate: `suite del repo y perfilado del flujo afectado`
- **MUST** No tratar React Compiler ni `useMemo` como una caché general para funciones, red o datos persistentes. Porqué: su alcance es la optimización de componentes y Hooks, no la semántica de almacenamiento de la aplicación. Gate: `revisión de ownership y lifetime de la caché`

### Testing y accesibilidad

- **MUST** Probar resultados observables mediante texto, roles e interacción en vez de estado interno o detalles privados. Porqué: el test debe describir el contrato del usuario y sobrevivir refactors legítimos. Gate: `runner de tests del repo con queries accesibles`
- **MUST** Cubrir los estados de carga, vacío, error, éxito y reintento que cambie el flujo. Porqué: el happy path no verifica cómo se recupera la interfaz ante respuestas normales del sistema. Gate: `tests de componentes o integración`
- **MUST** Usar elementos y atributos semánticos para que teclado y tecnologías asistivas puedan operar los controles. Porqué: un handler de click sobre un elemento visual no hereda rol, foco ni activación de teclado. Gate: `linter de accesibilidad y tests por rol`
- **MUST** No actualizar snapshots sin revisar su significado ni usarlos como única prueba de una interacción. Porqué: aceptar el diff mecánicamente no demuestra que el comportamiento sea correcto. Gate: `revisión de snapshots más assertions observables`

### Web en celular y webview

- **MUST** Mantener el viewport en `width=device-width, initial-scale=1`, sin `user-scalable=no` ni `maximum-scale=1`, y declararlo a mano con `<meta name="viewport">` solo si el framework no lo genera (Next.js lo emite solo). Porqué: bloquear el zoom impide ampliar el contenido a quien lo necesita y falla WCAG 1.4.4; un segundo `<meta>` duplica el tag. Gate: `grep de user-scalable y maximum-scale en el HTML base`
- **MUST** Dar a inputs, selects y textareas un `font-size` computado de al menos 16px en mobile. Porqué: Safari en iOS hace zoom automático al enfocar un campo con texto menor a 16px y desarma el layout. Gate: `prueba en Safari iOS o WebKit con viewport mobile`
- **MUST** Mantener visible el campo enfocado cuando aparece el teclado virtual, sin saltos de scroll al escribir: ajustar el layout con `interactive-widget=resizes-content` en Chromium y Android, y con la API `visualViewport` en Safari iOS, que no soporta `interactive-widget`; `scrollIntoView` queda solo como corrección puntual. Porqué: el teclado achica el viewport visual y un layout fijo puede tapar el input o saltar en cada tecla. Gate: `prueba manual o E2E con teclado virtual en Safari iOS y Chrome Android`
- **MUST** Respetar las zonas seguras en barras fijas y contenido a pantalla completa con `viewport-fit=cover` y `env(safe-area-inset-*)`. Porqué: notch, isla dinámica y barra de gestos tapan los controles pegados a los bordes, y sin `viewport-fit=cover` los insets valen 0. Gate: `prueba visual en un dispositivo o simulador con notch`
- **SHOULD** Dar a cada control táctil un área de al menos 44pt (44px CSS) con separación entre targets vecinos. Porqué: los targets chicos generan toques errados; 44pt es el mínimo de las guías de Apple y WCAG 2.5.8 exige al menos 24px. Gate: `auditoría de accesibilidad o medición en DevTools`
- **MUST** Ofrecer un camino táctil y de teclado para todo lo que se revela con `:hover` o eventos de mouse. Porqué: en pantallas táctiles no hay hover y la acción o la información quedan inaccesibles. Gate: `prueba con emulación táctil y navegación por teclado`

### Verificación

- **MUST** Ejecutar lint, typecheck, tests y build mediante los scripts existentes del repo. Porqué: Hooks, tipos, comportamiento y bundling fallan por mecanismos distintos. Gate: scripts de CI declarados en `package.json`
- **MUST** Para cambios de interfaz, verificar navegación, foco, estados async y comportamiento accesible en el entorno que integra React. Porqué: un test unitario no ejecuta necesariamente routing, CSS, plataforma ni lector de pantalla. Gate: `tests de integración o prueba humana definida por el proyecto`
- **MUST** Informar versiones relevantes, comandos ejecutados y flujos visuales o accesibles que quedaron sin probar. Porqué: la evidencia debe distinguir comprobación automática de validación pendiente. Gate: `reporte final` con comando y resultado

### Lectura ampliada

- [Rules of React](https://react.dev/reference/rules) — pureza, inmutabilidad y responsabilidades del render.
- [eslint-plugin-react-hooks](https://react.dev/reference/eslint-plugin-react-hooks) — reglas recomendadas y diagnósticos del compilador.
- [Rules of Hooks](https://react.dev/reference/eslint-plugin-react-hooks/lints/rules-of-hooks) — orden de Hooks y caso especial de `use`.
- [Choosing the State Structure](https://react.dev/learn/choosing-the-state-structure) — estado mínimo y fuentes de verdad.
- [Extracting State Logic into a Reducer](https://react.dev/learn/extracting-state-logic-into-a-reducer) — cuándo reunir transiciones.
- [You Might Not Need an Effect](https://react.dev/learn/you-might-not-need-an-effect) — derivados, eventos y sincronización externa.
- [Synchronizing with Effects](https://react.dev/learn/synchronizing-with-effects) — setup, cleanup y lifecycle.
- [Removing Effect Dependencies](https://react.dev/learn/removing-effect-dependencies) — dependencias reactivas sin supresiones.
- [Preserving and Resetting State](https://react.dev/learn/preserving-and-resetting-state) — posición, tipo y keys.
- [React Compiler: Introduction](https://react.dev/learn/react-compiler/introduction) — memoización automática y alcance.
- [useEffectEvent](https://react.dev/reference/react/useEffectEvent) — separar lógica no reactiva sin ocultar dependencias.
- [MDN: viewport meta](https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Elements/meta/name/viewport) — `width`, `initial-scale`, `interactive-widget` y por qué no bloquear el zoom.
- [MDN: env()](https://developer.mozilla.org/en-US/docs/Web/CSS/env) — `safe-area-inset-*` y `viewport-fit=cover`.
- [MDN: VisualViewport](https://developer.mozilla.org/en-US/docs/Web/API/VisualViewport) — viewport visual con teclado virtual.
- [WebKit: Designing Websites for iPhone X](https://webkit.org/blog/7929/designing-websites-for-iphone-x/) — zonas seguras en Safari.
- [Apple HIG: Accessibility](https://developer.apple.com/design/human-interface-guidelines/accessibility) — tamaño mínimo de 44pt para controles táctiles.
- [WCAG 2.2: Target Size (Minimum)](https://www.w3.org/WAI/WCAG22/Understanding/target-size-minimum.html) — mínimo de 24px y separación entre targets.

## Ajustes de este proyecto
<!-- coding-policies:ajustes:start -->
<!-- coding-policies:ajustes:end -->

# Лабораторная работа #5

![GitHub Classroom Workflow](../../workflows/GitHub%20Classroom%20Workflow/badge.svg?branch=master)

## OAuth2 Authorization

### Формулировка

На базе [Лабораторной работы #4](https://github.com/bmstu-rsoi/lab2-template) реализовать OAuth2 token-based
авторизацию.

* Для авторизации использовать OpenID Connect, в роли Identity Provider использовать стороннее решение.
* На Identity Provider настроить
  использование [Resource Owner Password flow](https://auth0.com/docs/authorization/flows/resource-owner-password-flow)
  (в одном запросе передается `clientId`, `clientSecret`, `username`, `password`).
* Все методы `/api/**` (кроме `/api/v1/authorize` и `/api/v1/callback`) на всех сервисах закрыть token-based
  авторизацией.
* В качестве токена использовать [JWT](https://jwt.io/introduction), для валидации токена
  использовать [JWKs](https://auth0.com/docs/security/tokens/json-web-tokens/json-web-key-sets), _запрос к Identity
  Provider делать не нужно_.
* JWT токен пробрасывать между сервисами, при получении запроса валидацию токена так же реализовать через JWKs.
* Убрать заголовок `X-User-Name` и получать пользователя из JWT-токена.
* Если авторизация некорректная (отсутствие токена, ошибка валидации JWT токена, закончилось время жизни токена
  (поле `exp` в payload)), то отдавать 401 ошибку.
* В `scope` достаточно указывать `openid profile email`.

### Требования

1. Для автоматических прогонов тестов в файле [autograding.json](.github/classroom/autograding.json)
   и [classroom.yml](.github/workflows/classroom.yml) заменить `<variant>` на ваш вариант.
1. Код хранить на Github, для сборки использовать Github Actions.
1. Каждый сервис должен быть завернут в docker.
1. В classroom.yml дописать шаги на сборку, прогон unit-тестов.

### Пояснения

1. В роли Identity Provider можно использовать любое решение, вот несколько рабочих вариантов:
    1. [Okta](https://developer.okta.com/docs/guides/)
    2. [Auth0](https://auth0.com/developers)
2. Для получения metadata для OpenID Connect можно
   использовать [Well-Known URI](https://auth0.com/docs/security/tokens/json-web-tokens/locate-json-web-key-sets):
   `https://[base-server-url]/.well-known/openid-configuration`.
3. Из Well-Known metadata можно получить Issuer URI и JWKs URI.
4. Для реализации OAuth2 можно использовать сторонние библиотеки.

### Прием задания

1. При получении задания у вас создается fork этого репозитория для вашего пользователя.
2. После того как все тесты успешно завершатся, в Github Classroom на Dashboard будет отмечено успешное выполнение
   тестов.

### Варианты заданий

Распределение вариантов заданий аналогично [ЛР #2](https://github.com/bmstu-rsoi/lab2-template).

## Реализация

Вариант API — `v1`. OpenID Connect provider — **Keycloak 26**. Он запускается
вместе с системой: в Docker Compose и отдельным Helm chart для EKS. Gateway предоставляет публичный endpoint
`POST /api/v1/authorize` с JSON-полями `username` и `password`; он выполняет
Resource Owner Password flow и возвращает ответ token endpoint без изменений.
Для всех прочих HTTP-маршрутов требуется заголовок `Authorization: Bearer
<access_token>`. Gateway передаёт его внутренним сервисам, где подпись, `iss`
и `exp` проверяются повторно по RSA-ключу из JWKs.

В Keycloak импортируется realm `rsoi`, confidential client `rsoi-gateway` с
включённым Direct Access Grants и тестовый пользователь `testkate` / `test`.
Claim `username` этого пользователя равен `Test Max`, поэтому сохранённые из
ЛР-4 данные остаются доступны. Для production замените значения паролей и
client secret в Helm values до первого запуска Keycloak (значение подставляется
и в импортируемый realm).

Настройка OIDC, которую сервисы получают через Kubernetes Secret `oidc`:

| Переменная | Назначение |
| --- | --- |
| `OIDC_ISSUER` | точное значение issuer из OIDC metadata |
| `OIDC_JWKS_URL` | URI набора публичных JWKs |
| `OIDC_TOKEN_URL` | token endpoint (нужен только gateway) |
| `OIDC_CLIENT_ID` / `OIDC_CLIENT_SECRET` | confidential client для password flow |
| `OIDC_USER_CLAIM` | claim с именем пользователя, по умолчанию `sub` |

Для EKS GitHub Actions разворачивает `postgres`, затем `keycloak`, после чего
сервисы. Локальный запуск не требует внешних OIDC-переменных: `docker compose up --build`.
Health check `/manage/health` намеренно не требует токена для Kubernetes probes.

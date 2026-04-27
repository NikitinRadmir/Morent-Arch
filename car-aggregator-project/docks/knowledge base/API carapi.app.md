# API

В качестве источника информации был отобран открытый API,
а именно ресурс [carapi.app](https://carapi.app/), который
бесплатно предоставляет информацию об автомобилях,
выпущенных в промежутке 2015 - 2020 годов.

Данное ограничение для проекта не является существенным,
так как на данный момент рынок автомобилей в России в
основном состоит из модельного ряда этих годов выпуска.

## Ограничения

- Информация об автомобилях выпуска с 2015 - 2020
- Не более 600 запросов в минуту

## Взаимодействие с API

Для взаимодействия с сервисом для начала необходимо
авторизоваться.

```shell
curl -X 'POST' \
     'https://carapi.app/api/auth/login'
     -H 'accept: text/plain' \
     -H 'Content-Type: application/json' \
     -d '{
     "api_token": "your_token"
     "api_secret": "your_secret"'
     }'
```

И в ответ получаем JWT токен, который мы в дальнейшем
используем в запросах. Например:

```shell
curl -X 'GET' \
  'https://carapi.app/api/trims/v2?limit=10&model=Range%20Rover' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer <your_jwt_token>'
```

```json
{
  "collection": {
    "url": "\/api\/trims\/v2?limit=10\u0026model=Range%20Rover",
    "count": 10,
    "pages": 6,
    "total": 57,
    "next": "\/api\/trims\/v2?model=Range+Rover\u0026page=2\u0026limit=10",
    "prev": "",
    "first": "\/api\/trims\/v2?limit=10\u0026model=Range%20Rover",
    "last": "\/api\/trims\/v2?model=Range+Rover\u0026page=6\u0026limit=10"
  },
  "data": [
    {
      "id": 22722,
      "make_id": 14,
      "model_id": 4021,
      "submodel_id": 64902,
      "year": 2015,
      "make": "Land Rover",
      "model": "Range Rover",
      "series": null,
      "submodel": "Autobiography",
      "trim": "Autobiography",
      "description": "Autobiography 4dr SUV 4WD (5.0L 8cyl S\/C 8A)",
      "msrp": 137995,
      "invoice": 126265,
      "created": "2023-06-29T21:05:54-04:00",
      "modified": "2023-06-29T21:05:54-04:00"
    },
    {
      "id": 22723,
      "make_id": 14,
      "model_id": 4021,
      "submodel_id": 64902,
      "year": 2015,
      "make": "Land Rover",
      "model": "Range Rover",
      "series": null,
      "submodel": "Autobiography",
      "trim": "Autobiography LWB",
      "description": "Autobiography LWB 4dr SUV 4WD (5.0L 8cyl S\/C 8A)",
      "msrp": 142995,
      "invoice": 130840,
      "created": "2023-06-29T21:05:54-04:00",
      "modified": "2023-06-29T21:05:54-04:00"
    },
    ...,
    {
      "id": 22720,
      "make_id": 14,
      "model_id": 4021,
      "submodel_id": 64905,
      "year": 2015,
      "make": "Land Rover",
      "model": "Range Rover",
      "series": null,
      "submodel": "Supercharged",
      "trim": "Supercharged Limited",
      "description": "Supercharged Limited 4dr SUV 4WD (5.0L 8cyl S\/C 8A)",
      "msrp": 118845,
      "invoice": 109668,
      "created": "2023-06-29T21:05:54-04:00",
      "modified": "2023-06-29T21:05:54-04:00"
    }
  ]
}
```

---

Также помимо этих данных, сервис взаимодействует по тегам:

- Years
- Makes
- Models
- Sub models
- Trims
- Vin Decoder
- License Plate
- Bodies
- Engines
- Mileages
- Colors (Exterior)
- Colors (Interior)
- OBD Codes
- Data Feeds
- Vehicle Attribute
- Account

> Подробнее ознакомиться с [документацией](https://carapi.app/api#/)
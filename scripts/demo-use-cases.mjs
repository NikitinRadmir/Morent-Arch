import { execFileSync } from 'node:child_process';

const API = process.env.MORENT_API_URL || 'http://localhost:1488';
const USER_API = process.env.USER_SYSTEM_API_URL || 'http://localhost:8082';
const ROOT = new URL('..', import.meta.url).pathname.replace(/^\/([A-Za-z]:)/, '$1');

class CookieJar {
  constructor() {
    this.cookies = new Map();
  }

  apply(headers = {}) {
    const cookie = [...this.cookies.entries()].map(([k, v]) => `${k}=${v}`).join('; ');
    return cookie ? { ...headers, Cookie: cookie } : headers;
  }

  store(response) {
    const values = response.headers.get('set-cookie');
    if (!values) return;
    for (const part of values.split(/,\s*(?=[^;=]+=[^;]+)/)) {
      const [pair] = part.split(';');
      const eq = pair.indexOf('=');
      if (eq > 0) {
        this.cookies.set(pair.slice(0, eq).trim(), pair.slice(eq + 1).trim());
      }
    }
  }
}

const jar = new CookieJar();

async function request(path, { method = 'GET', body, headers = {}, jar: requestJar = jar, base = API } = {}) {
  const response = await fetch(`${base}${path}`, {
    method,
    headers: requestJar.apply({
      ...(body === undefined ? {} : { 'Content-Type': 'application/json' }),
      ...headers,
    }),
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  requestJar.store(response);
  const text = await response.text();
  let data = null;
  if (text) {
    try {
      data = JSON.parse(text);
    } catch {
      data = text;
    }
  }
  return { status: response.status, data, text };
}

function assertCase(name, ok, details) {
  const mark = ok ? 'PASS' : 'FAIL';
  console.log(`${mark} ${name}`);
  if (details !== undefined) {
    console.log(JSON.stringify(details, null, 2));
  }
  if (!ok) {
    process.exitCode = 1;
  }
}

function getVerificationCode(email) {
  const sql = `select email_verification_code from users where email='${email.replaceAll("'", "''")}'`;
  return execFileSync('docker', [
    'compose',
    '--project-directory',
    `${ROOT}/Morent-project`,
    '-f',
    `${ROOT}/Morent-project/docker-compose.yml`,
    'exec',
    '-T',
    'db',
    'sh',
    '-lc',
    `psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -tAc "${sql}"`,
  ], { encoding: 'utf8' }).trim();
}

async function main() {
  const suffix = Date.now();
  const userEmail = `demo-${suffix}@example.com`;
  const userPassword = 'Demo123!AA';
  const userJar = new CookieJar();

  const badLogin = await request('/auth/login', {
    method: 'POST',
    body: { email: 'admin@morent.com', password: '123' },
    jar: userJar,
  });
  assertCase('UX: короткий пароль возвращает публичную ошибку', badLogin.status === 400 && !/LoginRequest|failed on the/.test(JSON.stringify(badLogin.data)), badLogin);

  const register = await request('/auth/register', {
    method: 'POST',
    body: { name: 'Demo User', email: userEmail, password: userPassword },
    jar: userJar,
  });
  assertCase('UC-01: регистрация пользователя', register.status === 201, { status: register.status, email: userEmail });

  const code = getVerificationCode(userEmail);
  const verify = await request('/auth/verify-email', {
    method: 'POST',
    body: { email: userEmail, code },
    jar: userJar,
  });
  assertCase('UC-01: подтверждение email и сессия', verify.status === 200 && verify.data?.user?.emailVerified === true, { status: verify.status, code });

  const userLogin = await request('/auth/login', {
    method: 'POST',
    body: { email: userEmail, password: userPassword },
    jar: userJar,
  });
  assertCase('UC-01: вход пользователя', userLogin.status === 200 && Boolean(userLogin.data?.token), { status: userLogin.status });

  const cars = await request('/cars', { jar: userJar });
  const car = Array.isArray(cars.data) ? cars.data[0] : null;
  assertCase('UC-02: каталог автомобилей доступен', cars.status === 200 && Boolean(car?.id), { status: cars.status, carId: car?.id });

  await request('/bank/session', { method: 'POST', body: {}, jar: userJar });
  await request('/bank/deposit', { method: 'POST', body: { amount: 5000 }, jar: userJar });

  const day = 90 + Math.floor(Math.random() * 30);
  const start = new Date(Date.now() + day * 86400000).toISOString().slice(0, 10);
  const end = new Date(Date.now() + (day + 2) * 86400000).toISOString().slice(0, 10);
  const rentalPayload = {
    carId: Number(car.id),
    startDate: start,
    endDate: end,
    totalPrice: Number(car.price) * 2,
  };

  const rental = await request('/rentals', { method: 'POST', body: rentalPayload, jar: userJar });
  assertCase(
    'TC-03 / UC-02: успешное создание аренды',
    rental.status === 201 && Number(rental.data?.car?.id) === Number(car.id),
    { status: rental.status, rental: rental.data },
  );

  const duplicate = await request('/rentals', { method: 'POST', body: rentalPayload, jar: userJar });
  assertCase('TC-04 / UC-02: запрет двойной аренды', duplicate.status === 409, { status: duplicate.status, error: duplicate.data });

  const favorite = await request('/favorites', { method: 'POST', body: { carId: Number(car.id) }, jar: userJar });
  const favoriteAgain = await request('/favorites', { method: 'POST', body: { carId: Number(car.id) }, jar: userJar });
  const favorites = await request('/favorites', { jar: userJar });
  const favoriteCount = Array.isArray(favorites.data) ? favorites.data.filter((item) => Number(item.id) === Number(car.id)).length : 0;
  assertCase('UC-03: избранное без дублей', favorite.status === 201 && favoriteAgain.status === 201 && favoriteCount === 1, { favoriteCount });

  const comment = await request('/comments', {
    method: 'POST',
    body: { carId: Number(car.id), rating: 5, description: 'Demo review after rental' },
    jar: userJar,
  });
  assertCase('UC-06: комментарий после аренды', comment.status === 201, { status: comment.status, commentId: comment.data?.id });

  const forbiddenAdmin = await request('/Admin/Users', { jar: userJar });
  assertCase('TC-23 / UC-07: обычный пользователь не проходит в админку', forbiddenAdmin.status === 403, { status: forbiddenAdmin.status, error: forbiddenAdmin.data });

  const adminLogin = await request('/auth/login', {
    method: 'POST',
    body: { email: 'admin@morent.com', password: 'admin123' },
  });
  assertCase('UC-07: вход администратора', adminLogin.status === 200 && adminLogin.data?.user?.role === 'admin', { status: adminLogin.status });

  const aggregator = await request('/Admin/Aggregator/Cars?q=toyota');
  const trim = aggregator.data?.cars?.[0];
  assertCase('TC-08 / UC-04: поиск в агрегаторе', aggregator.status === 200 && Boolean(trim), { status: aggregator.status, count: aggregator.data?.count });

  if (trim) {
    const imported = await request('/Admin/Aggregator/Import', {
      method: 'POST',
      body: { trim },
    });
    assertCase('TC-08 / UC-04: импорт авто из агрегатора', [200, 201].includes(imported.status) && Boolean(imported.data?.id), { status: imported.status, carId: imported.data?.id, name: imported.data?.name });
  }

  const usJar = new CookieJar();
  const usEmail = `us-${suffix}@example.com`;
  const usRegister = await request('/api/auth/register', {
    method: 'POST',
    base: USER_API,
    jar: usJar,
    body: {
      email: usEmail,
      password: 'Password123!',
      first_name: 'Role',
      last_name: 'Demo',
      company_name: `DemoCo${suffix}`,
      department: 'IT',
      position: 'Admin',
    },
  });
  const usLogin = await request('/api/auth/login', {
    method: 'POST',
    base: USER_API,
    jar: usJar,
    body: { email: usEmail, password: 'Password123!' },
  });
  const authHeaders = { Authorization: `Bearer ${usLogin.data?.access_token || ''}` };
  const companies = await request('/api/companies', { base: USER_API, jar: usJar, headers: authHeaders });
  const company = companies.data?.data?.find((item) => item.name === `DemoCo${suffix}`) || companies.data?.data?.[0];
  const role = await request('/api/roles', {
    method: 'POST',
    base: USER_API,
    jar: usJar,
    headers: authHeaders,
    body: { company_id: company?.id, name: `demo-role-${suffix}`, description: 'demo role' },
  });
  assertCase(
    'UC-08: user-system создает роль в компании',
    usRegister.status === 201 && usLogin.status === 200 && role.status === 201,
    { register: usRegister.status, login: usLogin.status, companyId: company?.id, role: role.data },
  );

  console.log('\nDemo finished. Use --with-aggregator-outage manually: stop aggregator, call /Admin/Aggregator/Cars, expect 503, start aggregator again.');
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});

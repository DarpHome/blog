export class RestError extends Error {
  constructor(res) {
    super(res.message);
    this.name = 'RestError';
    this.code = res.code;
    if ('extra' in res) {
      this.extra = res.extra;
    } else {
      this.extra = null;
    }
    if ('fields' in res) {
      this.fields = Object.fromEntries(Object.entries(res.fields).map(p => [p[0], new RestError(p[1])]));
    } else {
      this.fields = null;
    }
  }
}

export class Unauthorized extends Error {
  constructor(res) {
    super(res);
    this.name = 'Unauthorized';
  }
}

export class Forbidden extends Error {
  constructor(res) {
    super(res);
    this.name = 'Forbidden';
  }
}

export class NotFound extends Error {
  constructor(res) {
    super(res);
    this.name = 'NotFound';
  }
}

const EPOCH = '1699567200000';

export class Snowflaked {
  static epoch() {
    return EPOCH;
  }
  constructor(id) {
    if (typeof id !== 'string') {
      throw new TypeError('id must be string');
    }
    this.id = id;
  }
  get createdAt() {
    return new Date(Number((BigInt(this.id) >> BigInt(22)) + BigInt(EPOCH)));
  }
}

export class Bitset {
  constructor(value) {
    this.value = value;
  }
  hasAll(flags) {
    return (this.value & flags) === flags;
  }
  hasAny(flags) {
    return (this.value & flags) !== 0;
  }
  has(bit) {
    return this.hasAny(1 << bit);
  }
  toggle(flags) {
    this.value &= ~flags;
  }
  set(flags, value) {
    if (value) this.value |= flags;
    else this.value ^= flags;
  }
}

export class UserFlags extends Bitset {
  constructor(value) {
    super(value);
  }
  get deleted() {
    return this.has(0);
  }
  get staff() {
    return this.has(1);
  }
  get bot() {
    return this.has(2);
  }
}

export class User extends Snowflaked {
  constructor(data) {
    super(data.id);
    this.avatar = data.avatar;
    this.bio = data.bio;
    this.flags = new UserFlags(data.flags);
    if ('username' in data)
      this.username = data.username;
    else
      this.username = null;
  }
}

export class RegisterResult {
  constructor(data) {
    this.token = data.token;
    this.user = new User(data.user);
  }
}

export class Client {
  constructor(token = null, options = {}) {
    this.token = token || null;
    options ||= {};
    if ('base' in options)
      this.base = options.base;
    else
      this.base = '/api/v1';
  }
  withToken(token) {
    this.token = token;
    return this;
  }
  withoutToken() {
    return this.withToken(null);
  }
  async request(method, route, options = {}) {
    const headers = {};
    if (this.token) headers['Authorization'] = this.token;
    if ('headers' in options) {
      headers = Object.assign(headers, options.headers);
      delete options.headers;
    }
    if ('body' in options && typeof options.body !== 'string') {
      headers['Content-Type'] = 'application/json';
      options.body = JSON.stringify(options.body);
    }
    const r = await fetch(this.base + route, {
      headers,
      method,
      ...options,
    });
    if (r.status >= 400 && r.status < 600) {
      const j = await r.json();
      if (r.status === 401)
        throw new Unauthorized(j);
      if (r.status === 403)
        throw new Forbidden(j);
      if (r.status === 404)
        throw new NotFound(j);
      throw new RestError(j);
    }
    return r;
  }
  async register({username, password}) {
    const r = await this.request('POST', '/auth/register', {
      body: {
        username,
        password,
      },
    });
    return new RegisterResult(await r.json());
  }
  async login({username, password}) {
    const r = await this.request('POST', '/auth/login', {
      body: {
        username,
        password,
      },
    });
    return (await r.json()).token;
  }
  async user(id = null) {
    if (id) {
      return new User(await (await this.request('GET', `/users/${id}`)).json());
    } else {
      return new User(await (await this.request('GET', `/users/@me`)).json());
    }
  }
}

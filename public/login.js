import * as api from './api.js';

document.getElementById('login').addEventListener('click', async (ev) => {
  const client = new api.Client();
  const r = document.getElementById('rest-response');
  r.innerText = '';
  document.getElementById('error-username').innerText = '';
  document.getElementById('error-password').innerText = '';
  try {
    const rr = await client.login({
      username: document.getElementById('username').value,
      password: document.getElementById('password').value,
    });

    try {
      localStorage.setItem('token', rr.token);
      localStorage.setItem('user', JSON.stringify(rr.user));
    } catch (err) {
      
    }
    r.style.color = '#10dd10';
    r.innerText = 'OK';
    window.location = '/profile';
  } catch (err) {
    console.log(err);
    r.style.color = '#dd1010';
    r.innerText = err.message;
    for (const [k, v] of err.fields ?? {}) {
      const table = {username: 'error-username', password: 'error-password'};
      if (k in table) {
        document.getElementById(table[k]).innerText = v.message;
      }
    }
  }
});
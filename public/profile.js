import * as api from './api.js';

function format(d) {
  const months = [
    'Jan', // 1
    'Feb', // 2
    'Mar', // 3
    'Apr', // 4
    'May', // 5
    'Jun', // 6
    'Jul', // 7
    'Aug', // 8
    'Sep', // 9
    'Oct', // 10
    'Nov', // 11
    'Dec', // 12
  ];
  return `${d.getDate()} ${months[d.getMonth()]} ${d.getFullYear()}`;
}

function sep() {
  const a = document.createElement('a');
  a.innerText = ' | ';
  return a;
}
function getBadges(flags) {
  const r = [];
  if (flags.staff) {
    const a = document.createElement('a');
    a.style.color = '#365dfd';
    a.innerText = 'STAFF';
    r.push(a);
  }
  if (flags.bot) {
    if (r.length) r.push(sep());
    const a = document.createElement('a');
    a.style.color = '#6263f0';
    a.innerText = 'BOT';
    r.push(a);
  }
  return r;
}

export async function pageLoaded() {
  const token = localStorage.getItem('token');
  if (!token) {
    alert('Please login or register to see your profile');
    window.location.href = '/auth/login';
    return;
  }
  const client = new api.Client(token);
  let user;
  try {
    user = await client.user();
  } catch (err) {
    if (err instanceof api.Unauthorized) {
      localStorage.removeItem('token');
      alert('Your token expired. Please login again');
      window.location.href = '/auth/login';
    } else {
      console.error('internal error');
      console.error(err);
      debugger;
    }
    return;
  }
  /*        <h1>Profile</h1>
  <br/>
  Username: <a id="username"></a> <span id="badges"></span><br/><br/>
  <div id="bio">
  </div>*/
  document.getElementById('username').innerText = user.username;
  const badges = document.getElementById('badges');
  badges.innerText = '';
  badges.append(...getBadges(user.flags));
  document.getElementById('bio').innerText = user.bio;
  document.getElementById('joindate').innerText = format(user.createdAt);
};
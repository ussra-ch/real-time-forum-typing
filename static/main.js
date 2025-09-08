import { logindiv } from "./login.js";
import { login } from "./login.js";
import { errorPage } from "./errorPage.js";


export function main() {
  if (document.getElementById('style')) {
    document.getElementById('style').remove()
  }
  const currentUrl = window.location.href;
  const urlArr = currentUrl.split('/')

  if (urlArr[urlArr.length - 1] != "" || urlArr.length != 4) {
    errorPage('404')
    return
  }
  const html = document.querySelector('html');
  //html.style.filter = 'blur(30px)';
  setTimeout(() => {
    html.style.filter = 'none';
  }, 700);

  logindiv();
  login()
}


main()

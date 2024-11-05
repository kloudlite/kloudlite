<a name="readme-top"></a>

<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="https://github.com/kloudlite/web">
    <img src="static/common/logo-with-name.png" alt="Logo" width="auto" height="30">
  </a>

  <h3 align="center">Kloudlite Web App</h3>

  <p align="center">
    All web apps of the kloudlite.io is maintained from this repo.
    <br />
    <a href="https://github.com/kloudlite/web/issues">Report Bug</a>
    ·
    <a href="https://github.com/kloudlite/web/issues">Request Feature</a>
  </p>
</div>

<br />

[![Product Name Screen Shot][product-screenshot]](https://kloudlite.io)

<br />


<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#contributing">Contributing</a></li>
  </ol>
</details>



<!-- ABOUT THE PROJECT -->
<!-- ## About The Project -->



### Built With

As this is web app. it's built on html, css and Javascript. all the frameworks, libraries used in this project listed below.

* [![Remix][Remix.Run]][Remix-url]
* [![React][React.js]][React-url]
* [![GraphQL][GraphQL]][GraphQL-url]
* [![TailwindCSS][Tailwind.CSS]][Tailwind-url]

> For more information about the libraries and frameworks used in this project. you can visit the `package.json` file.

---

> And this repo contains all the web apps. like: auth, console, accounts, socket, website.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- GETTING STARTED -->
## Getting Started

### Prerequisites

* npm
* pnpm
* taskgo
* docker


### Installation

1. Clone the repo
   ```sh
   git clone git@github.com:kloudlite/web.git
   ```
2. Install NPM packages
   ```sh
   pnpm install
   ```
<p align="right">(<a href="#readme-top">back to top</a>)</p>


<!-- USAGE EXAMPLES -->
## Usage

To start the application in development mode. simply execute the command:

```sh
task app={app-name} tscheck={yes/no}
```

To Build the application. simply execute the command:

```sh
task build app={app-name}
```

To Run the application. simply execute the command:

```sh
task run app={app-name}
```

To Build Docker Image of the application. simply execute the command:

```sh
task docker-build app={app-name} tag={tag-name}
```


> visit the `package.json` file or `Taskfile.yaml` file for more information about the commands.


<p align="right">(<a href="#readme-top">back to top</a>)</p>


<!-- CONTRIBUTING -->
## Contributing

1. Fork the Project
2. Clone the Project
3. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
4. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
5. Push to the Branch (`git push origin feature/AmazingFeature`)
6. Open a Pull Request

<p align="right">(<a href="#readme-top">back to top</a>)</p>


<!-- MARKDOWN LINKS & IMAGES -->
[product-screenshot]: ./static/common/screenshot.png
[React.js]: https://img.shields.io/badge/React-20232A?style=for-the-badge&logo=react&logoColor=61DAFB
[React-url]: https://reactjs.org/
[Remix.Run]: https://img.shields.io/badge/Remix-20232A?style=for-the-badge&logo=remix&logoColor=be123c
[Remix-url]: https://remix.run/
[GraphQl]: https://img.shields.io/badge/GraphQl-20232A?style=for-the-badge&logo=graphql&logoColor=E10098
[GraphQl-url]: https://graphql.org/
[Tailwind.CSS]: https://img.shields.io/badge/Tailwind-20232A?style=for-the-badge&logo=tailwindcss&logoColor=38bdf8
[Tailwind-url]: https://tailwindcss.com/

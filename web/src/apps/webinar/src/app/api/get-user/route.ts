// import axios from 'axios';
// import { cookies } from 'next/headers';
// import { NextResponse } from 'next/server';

import { NextResponse } from "next/server";


// // To handle a GET request to /api
// export async function GET(request) {
//     // Do whatever you want
//     const cookie = cookies().get("hotspot-session")
//     console.log("cookies ++++++>>>>>>>>>>>>>>",
//         cookies().get("hotspot-session"), "lkkk", cookie?.value);
//     const res = await axios({
//         url: "https://auth1.dev.kloudlite.io/api",
//         method: 'post',
//         // withCredentials: true,
//         data: {
//             method: 'whoAmI',
//             args: [{}],
//         },
//         headers: {
//             'Content-Type': 'application/json; charset=utf-8',
//             connection: 'keep-alive',
//             cookie: 'hotspot-session=' + cookie?.value + ';',
//         },
//     });
//     console.log("++++++>>>>>>>>>>>>>> response", res);
//     return NextResponse.json(res.data);

// }

// To handle a POST request to /api
export async function POST() {
    // Do whatever you want
    return NextResponse.json({ message: "Hello World" }, { status: 200 });
}
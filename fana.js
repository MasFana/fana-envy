const { APP_ENV, DB_HOST, DEBUG } = process.env;
console.log("ENVs:", { APP_ENV, DB_HOST, DEBUG });

setInterval(() => {
    console.log("Hello JS");
}, 3000);


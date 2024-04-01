import Database from "better-sqlite3";
import puppeteer from "puppeteer-core";

async function main() {
  const db = new Database("../db.db");
  db.pragma("journal_mode = WAL");

  db.exec("delete from laptops");
  db.exec("delete from cpus");

  const laptopStmt = db.prepare(
    `INSERT INTO laptops (title, link, price, description, model, searchtext)  
   VALUES (@title, @link, @price, @description, @model, @searchtext)`
  );

  const cpuStmt = db.prepare(
    `INSERT INTO cpus (family, model, cores, threads, clockspeed, l3cache, gpucores, gpuclockspeed, tdp) 
   VALUES (@family, @model, @cores, @threads, @clockspeed, @l3cache, @gpucores, @gpuclockspeed, @tdp)`
  );

  const browser = await puppeteer.launch({ executablePath: "/usr/bin/chromium-browser", headless: true });
  const page = await browser.newPage();

  await page.goto(`file:///home/user/Documents/ryzenlaptops/app/scraper/cpus.html`);

  const cpus = await page.$$eval("table tbody tr", (rows) =>
    rows.map((row) => {
      const cells = row.querySelectorAll("td");
      return {
        family: cells[0].innerText,
        model: cells[1].innerText,
        cores: Number(cells[2].innerText),
        threads: Number(cells[3].innerText),
        clockspeed: Number(cells[4].innerText),
        l3cache: Number(cells[5].innerText),
        gpucores: Number(cells[6].innerText),
        gpuclockspeed: Number(cells[7].innerText),
        tdp: Number(cells[8].innerText),
      };
    })
  );

  for (const element of cpus) {
    cpuStmt.run({
      family: element.family,
      model: element.model,
      cores: element.cores,
      threads: element.threads,
      clockspeed: element.clockspeed,
      l3cache: element.l3cache,
      gpucores: element.gpucores,
      gpuclockspeed: element.gpuclockspeed,
      tdp: element.tdp,
    });
  }

  for (let i = 1; i < 5; i++) {
    await page.goto(`https://www.kleinanzeigen.de/s-notebooks/anbieter:privat/seite:${i}/ryzen/k0c278`, {
      waitUntil: "domcontentloaded",
    });

    const laptops = await page.$$eval("#srchrslt-adtable > li", (items) =>
      items.map((item) => {
        let title = "";
        let link = "";
        let price = 0;
        let description = "";
        let model = "";
        let text = "";

        const anchorTag = item.querySelector("article > div.aditem-main > div.aditem-main--middle > h2 > a");
        const priceTag = item.querySelector(
          "article > div.aditem-main > div.aditem-main--middle > div.aditem-main--middle--price-shipping > p"
        );

        if (anchorTag instanceof HTMLAnchorElement && priceTag instanceof HTMLParagraphElement) {
          title = anchorTag.innerText.toLowerCase();
          link = anchorTag.href;
          price = Number(priceTag?.innerText?.split(" ")[0].replace("VB", "0").replace(".", ""));
        }

        return { title, link, price, description, model, text };
      })
    );

    for (const element of laptops) {
      try {
        await page.goto(element.link, { waitUntil: "domcontentloaded" });
        element.description = await page.$eval("#viewad-description-text", (el) => {
          if (el instanceof HTMLParagraphElement) return el.innerText?.toLowerCase().replace(/\s/g, " ");
          else return "";
        });
      } catch (error) {
      } finally {
        element.text = element.title + " " + element.description.replace(/\b(\d{4})\b/g, "$1u");
        element.model = cpus.find((el) => element.text.includes(el.model))?.model || "";

        if (element.title)
          laptopStmt.run({
            title: element.title,
            link: element.link,
            price: element.price,
            description: element.description,
            model: element.model,
            searchtext: element.text,
          });
      }
    }
  }

  await browser.close();
}
const start = performance.now();
await main();
const end = performance.now();
const result = end - start;
console.log(`delta time: ${result}ms`); 
                          

// Pruebas de navegador de apps/web en modo demo (datos ficticios, sin red). Variables: APP_URL, CHROME.
import test, { after, before } from "node:test";
import assert from "node:assert/strict";
import { createRequire } from "node:module";
import { readFileSync } from "node:fs";
import { chromium } from "playwright-core";

const APP_URL = process.env.APP_URL ?? "http://127.0.0.1:7703";
const require = createRequire(import.meta.url);
const AXE = readFileSync(require.resolve("axe-core/axe.min.js"), "utf8");
let browser;
before(async () => { browser = await chromium.launch({ executablePath: process.env.CHROME ?? "/usr/bin/google-chrome", args: ["--no-sandbox"] }); });
after(async () => browser?.close());

async function open({ viewport = { width: 1100, height: 800 }, colorScheme = "light" } = {}) {
  const context = await browser.newContext({ viewport, colorScheme });
  const page = await context.newPage();
  const problems = [];
  page.on("console", (m) => { if (m.type() === "error") problems.push(`console: ${m.text()}`); });
  page.on("pageerror", (e) => problems.push(`pageerror: ${e}`));
  page.on("requestfailed", (r) => problems.push(`requestfailed: ${r.url()} ${r.failure()?.errorText}`));
  await page.goto(APP_URL);
  return { page, context, problems };
}
const enter = async (page) => {
  await page.getByRole("button", { name: "Entrar" }).click();
  await page.getByRole("heading", { name: "Sofía" }).waitFor();
};
const say = async (page, text) => {
  await page.getByRole("textbox", { name: "Mensaje" }).fill(text);
  await page.getByRole("button", { name: "Enviar" }).click();
};
async function axe(page, label) {
  await page.evaluate(AXE);
  const result = await page.evaluate(() => axe.run(document, { runOnly: ["wcag2a", "wcag2aa", "best-practice"] }));
  const serious = result.violations.filter((v) => ["serious", "critical"].includes(v.impact));
  assert.deepEqual(serious.map((v) => `${v.id}: ${v.nodes.map((n) => n.target.join(" ")).join(" | ")}`), [], `axe (${label})`);
}
const finish = async ({ context, problems }) => { await context.close(); assert.deepEqual(problems, [], "errores de consola/red"); };

test("demo: acceso, bienvenida y salida", async () => {
  const s = await open();
  await s.page.getByText("Demo local · respuestas y acciones simuladas.").waitFor();
  await axe(s.page, "login");
  await enter(s.page);
  await s.page.getByText("Hola, Alex. ¿Qué te gustaría organizar hoy?").waitFor();
  await axe(s.page, "chat");
  await s.page.getByRole("button", { name: "Salir" }).click();
  await s.page.getByRole("button", { name: "Entrar" }).waitFor();
  await finish(s);
});

test("demo: conversación completa hasta una propuesta de acción (sólo lectura)", async () => {
  const s = await open();
  await enter(s.page);
  await say(s.page, "Quiero una pausa");
  await s.page.getByText("¿En qué momento del día te queda mejor?").waitFor();
  const user = s.page.locator(".msg.user");
  assert.equal(await user.count(), 1);
  assert.equal(await s.page.getByRole("textbox", { name: "Mensaje" }).inputValue(), "", "el campo se vacía al enviar");
  await s.page.getByRole("button", { name: "Por la tarde" }).click();
  await s.page.getByText("¿Cuánto tiempo te gustaría reservar?").waitFor();
  await s.page.getByRole("button", { name: "20 minutos" }).click();
  await s.page.getByText("Con lo que me contaste, te propongo esto:").waitFor();
  await s.page.getByText("Propuesta pendiente: se confirma desde la app móvil.").first().waitFor();
  assert.equal(await s.page.getByRole("button", { name: "Aprobar" }).count(), 0, "la web no confirma acciones");
  await finish(s);
});

test("mensajes en orden cronológico y sin mensajes vacíos", async () => {
  const s = await open();
  await enter(s.page);
  await say(s.page, "   ");
  assert.equal(await s.page.locator(".msg.user").count(), 0);
  await say(s.page, "Quiero ordenar la tarde");
  await s.page.getByText("¿En qué momento del día te queda mejor?").waitFor();
  const texts = await s.page.locator(".msg").evaluateAll((els) => els.map((e) => e.firstChild.textContent));
  assert.deepEqual(texts, ["Hola, Alex. ¿Qué te gustaría organizar hoy?", "Quiero ordenar la tarde", "¿En qué momento del día te queda mejor?"]);
  await finish(s);
});

test("teléfono 390 px: sin desbordamiento horizontal y accesible", async () => {
  const s = await open({ viewport: { width: 390, height: 844 } });
  await enter(s.page);
  await say(s.page, "Quiero una pausa");
  await s.page.getByText("¿En qué momento del día te queda mejor?").waitFor();
  assert.ok(await s.page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth), "desborda horizontalmente");
  await axe(s.page, "chat 390");
  await finish(s);
});

test("tema oscuro: contraste accesible en login y conversación", async () => {
  const s = await open({ colorScheme: "dark" });
  await axe(s.page, "login oscuro");
  await enter(s.page);
  await say(s.page, "Quiero una pausa");
  await s.page.getByText("¿En qué momento del día te queda mejor?").waitFor();
  await axe(s.page, "chat oscuro");
  await finish(s);
});

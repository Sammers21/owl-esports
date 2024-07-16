const delay = (ms) => new Promise((res) => setTimeout(res, ms));

function updateClipboard(newClip) {
  navigator.clipboard.writeText(newClip).then(
    () => {
      console.log("Clipboard updated");
    },
    () => {
      console.log("Clipboard failed to update");
    }
  );
}

async function sendPickToServer(tg, line) {
  line = line.replaceAll(" ", "_");
  json = {
    line: line,
    tg: tg,
  };
  console.log("Json sending:", json);
  const response = await fetch(
    "https://pvpq.net/owl-esports/pickline?" +
      new URLSearchParams(json).toString()
  );
  const data = await response.json();
  console.log("Response from pick server " + data);
}

async function main() {
  await delay(1000);
  // get tab name
  let tabName = document.title;
  console.log("Owl triggered, tab name:'" + tabName + "'");
  if (tabName !== "Dota2 Scoreboard") {
    return;
  }
  console.log("Owl is parsing the page...");

  const leftTeam = document.evaluate(
    "/html/body/div/div[3]/div/div[1]/div[1]",
    document,
    null,
    XPathResult.ANY_TYPE,
    null
  );

  const rightTeam = document.evaluate(
    "/html/body/div/div[3]/div/div[1]/div[3]",
    document,
    null,
    XPathResult.ANY_TYPE,
    null
  );

  const leftHeroes = document.evaluate(
    "/html/body/div/div[3]/div/div[3]/div[1]",
    document,
    null,
    XPathResult.ANY_TYPE,
    null
  );

  const rightHeroes = document.evaluate(
    "/html/body/div/div[3]/div/div[3]/div[3]",
    document,
    null,
    XPathResult.ANY_TYPE,
    null
  );

  function getChildNodes(node) {
    let childNodes = node.childNodes;
    let childNodesArray = Array.from(childNodes);
    return childNodesArray;
  }

  function getHeroFromNode(node) {
    let img = node.childNodes[0].childNodes[0];
    let src = img.getAttribute("src");
    let png = src.split("/").pop();
    let hero = png.split(".")[0];
    let replaceDash = hero.replace(/_/g, " ");
    console.log("Hero: ", replaceDash);
    if (replaceDash === "doom bringer") {
      replaceDash = "doom";
    } else if (replaceDash === "windrunner") {
      replaceDash = "windranger";
    } else if (replaceDash === "treant") {
      replaceDash = "treant protector";
    } else if (replaceDash === "shredder") {
      replaceDash = "timbersaw";
    } else if (replaceDash === "nevermore") {
      replaceDash = "shadow fiend";
    }
    return replaceDash;
  }

  function heroListToCommand(heroes) {
    return heroes.join(",");
  }

  function parseTeam(team, left) {
    let chindx = left ? 1 : 0;
    let childNodes = getChildNodes(getChildNodes(team)[chindx])[0];
    let teamName = childNodes.textContent;
    let ctag = childNodes.getAttribute("class");
    let isRadiant = ctag.includes("radiant");
    console.log("Team Name:", childNodes.textContent, "Is Radiant:", isRadiant);
    return { team: teamName, isRadiant: isRadiant };
  }

  try {
    let lt = leftTeam.iterateNext();
    let rt = rightTeam.iterateNext();
    console.log("Left Team: ", lt);
    console.log("Right Team: ", rt);
    let leftTeamParsed = parseTeam(lt, true);
    let rightTeamParsed = parseTeam(rt, false);
    if (leftTeamParsed.isRadiant) {
      radiantNode = leftHeroes.iterateNext();
      direNode = rightHeroes.iterateNext();
    } else {
      radiantNode = rightHeroes.iterateNext();
      direNode = leftHeroes.iterateNext();
    }
    console.log("Radiant Node: ", radiantNode);
    console.log("Dire Node: ", direNode);
    let childNodes = getChildNodes(radiantNode);
    let heroes = [];
    for (let i = 0; i < childNodes.length; i++) {
      let hero = getHeroFromNode(childNodes[i]);
      heroes.push(hero);
    }
    let direChildNodes = getChildNodes(direNode);
    for (let i = 0; i < direChildNodes.length; i++) {
      let hero = getHeroFromNode(direChildNodes[i]);
      heroes.push(hero);
    }
    console.log("All Heroes: ", heroes);
    let command = heroListToCommand(heroes).trim();
    console.log("Command: ", command);
    let team = leftTeamParsed.isRadiant ? leftTeamParsed.team + " vs " + rightTeamParsed.team : rightTeamParsed.team + " vs " + leftTeamParsed.team;
    console.log("Match: ", team);
    sendPickToServer(77107633, command);
  } catch (e) {
    console.log(e);
  }
}

main();

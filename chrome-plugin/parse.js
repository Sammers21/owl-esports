const delay = (ms) => new Promise((res) => setTimeout(res, ms));

function updateClipboard(newClip) {
  navigator.clipboard.writeText(newClip).then(
    () => {
      console.log("Clipboard updated");
    },
    () => {
      console.log("Clipboard failed to update");
    },
  );
}

async function sendPickToServer(pick) {
  const response = await fetch("http://localhost:3000/pick", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      pick: pick,
    }),
  });
  const data = await response.json();
  console.log(data);
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

  const radiant = document.evaluate(
    "/html/body/div/div[3]/div/div[3]/div[3]",
    document,
    null,
    XPathResult.ANY_TYPE,
    null
  );

  const dire = document.evaluate(
    "/html/body/div/div[3]/div/div[3]/div[1]",
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

  try {
    let radiantNode = radiant.iterateNext();
    let direNode = dire.iterateNext();
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
    let command = heroListToCommand(heroes);
    console.log("Command: ", command);
    updateClipboard(command);
  } catch (e) {
    console.log(e);
  }
}

main();

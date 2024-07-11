const radiant = document.evaluate(
  '/html/body/div/div[3]/div/div[3]/div[1]',
  document,
  null,
  XPathResult.ANY_TYPE,
  null
);

const dire = document.evaluate(
  '/html/body/div/div[3]/div/div[3]/div[3]',
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
  let src = img.getAttribute('src');
  let png = src.split('/').pop();
  let hero = png.split('.')[0];
  let replaceDash = hero.replace(/_/g, ' ');
  console.log('Hero: ', replaceDash);
  if (replaceDash === 'doom bringer') {
    replaceDash = 'doom';
  } else if (replaceDash === 'windrunner') {
    replaceDash = 'windranger';
  } else if (replaceDash === 'treant') {
    replaceDash = 'treant protector';
  }
  return replaceDash;
}

function heroListToCommand(heroes) {
  return heroes.join(',');
}

try {
  let radiantNode = radiant.iterateNext();
  let direNode = dire.iterateNext();
  console.log('Radiant Node: ', radiantNode);
  console.log('Dire Node: ', direNode);
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
  console.log('All Heroes: ', heroes);
  let command = heroListToCommand(heroes);
  console.log('Command: ', command);
} catch (e) {
  console.log(e);
}

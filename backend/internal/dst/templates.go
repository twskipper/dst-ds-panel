package dst

const DefaultClusterIni = `[GAMEPLAY]
game_mode = survival
max_players = 6
pvp = false

[NETWORK]
cluster_name = My DST Server
cluster_description =
cluster_password =
cluster_intention = cooperative

[MISC]
console_enabled = true

[SHARD]
shard_enabled = true
bind_ip = 127.0.0.1
master_ip = 127.0.0.1
master_port = 10888
cluster_key = defaultkey
`

const DefaultMasterServerIni = `[NETWORK]
server_port = 10999

[SHARD]
is_master = true

[STEAM]
master_server_port = 27018
authentication_port = 8768
`

const DefaultCavesServerIni = `[NETWORK]
server_port = 10998

[SHARD]
is_master = false
name = Caves

[STEAM]
master_server_port = 27019
authentication_port = 8769
`

const DefaultMasterLevelData = `return {
  desc="The standard Don't Starve experience.",
  hideminimap=false,
  id="SURVIVAL_TOGETHER",
  location="forest",
  max_playlist_position=999,
  min_playlist_position=0,
  name="Survival",
  numrandom_set_pieces=4,
  override_level_string=false,
  overrides={
    alternatehunt="default",
    angrybees="default",
    antliontribute="default",
    autumn="default",
    bearger="default",
    beefalo="default",
    beefaloheat="default",
    beequeen="default",
    bees="default",
    berrybush="default",
    birds="default",
    boons="default",
    butterfly="default",
    buzzard="default",
    cactus="default",
    carrot="default",
    catcoon="default",
    chess="default",
    crabking="default",
    darkness="default",
    day="default",
    deciduousmonster="default",
    deerclops="default",
    dragonfly="default",
    flint="default",
    flowers="default",
    frograin="default",
    frogs="default",
    ghostenabled="always",
    ghostsanitydrain="always",
    goosemoose="default",
    grass="default",
    has_ocean=true,
    healthpenalty="always",
    hounds="default",
    hunger="default",
    hunt="default",
    klaus="default",
    krampus="default",
    lightning="default",
    lightninggoat="default",
    lureplants="default",
    malbatross="default",
    merm="default",
    meteorshowers="default",
    moles="default",
    mosquitos="default",
    mushroom="default",
    penguins="default",
    perd="default",
    pigs="default",
    ponds="default",
    portalresurection="none",
    rabbits="default",
    reeds="default",
    regrowth="default",
    resettime="none",
    rock="default",
    rock_ice="default",
    sapling="default",
    season_start="default",
    shadowcreatures="default",
    spiderqueen="default",
    spiders="default",
    spring="default",
    summer="default",
    tallbirds="default",
    tentacles="default",
    touchstone="default",
    trees="default",
    tumbleweed="default",
    walrus="default",
    weather="default",
    wildfires="default",
    winter="default",
    world_size="default",
    task_set="default",
    start_location="default",
    layout_mode="LinkNodesByKeys",
    roads="default",
    loop="default",
    branching="default",
    spawnmode="fixed",
    keep_disconnected_tiles=true,
    no_joining_islands=true,
    no_wormholes_to_disconnected_tiles=true,
    has_ocean=true,
    wormhole_prefab="wormhole"
  },
  random_set_pieces={
    "Sculptures_2",
    "Sculptures_3",
    "Sculptures_4",
    "Sculptures_5",
    "Chessy_1",
    "Chessy_2",
    "Chessy_3",
    "Chessy_4",
    "Chessy_5",
    "Chessy_6"
  },
  required_prefabs={ "multiplayer_portal" },
  substitutes={  },
  version=4
}
`

const DefaultCavesLevelData = `return {
  background_node_range={ 0, 1 },
  desc="Caves!",
  hideminimap=false,
  id="DST_CAVE",
  location="cave",
  max_playlist_position=999,
  min_playlist_position=0,
  name="Caves",
  numrandom_set_pieces=0,
  override_level_string=false,
  overrides={
    banana="default",
    bats="default",
    berrybush="default",
    bunnymen="default",
    cave_ponds="default",
    cave_spiders="default",
    cavelight="default",
    chess="default",
    darkness="default",
    day="default",
    earthquakes="default",
    fern="default",
    fissure="default",
    flint="default",
    flower_cave="default",
    ghostenabled="always",
    ghostsanitydrain="always",
    grass="default",
    healthpenalty="always",
    hunger="default",
    lichen="default",
    marshbush="default",
    monkey="default",
    mushroom="default",
    mushtree="default",
    mushtree_moon="default",
    pillar_cave="default",
    portalresurection="none",
    reeds="default",
    regrowth="default",
    resettime="none",
    rock="default",
    rocky="default",
    sapling="default",
    season_start="default",
    slurper="default",
    slurtles="default",
    spiders="default",
    tentacles="default",
    touchstone="default",
    trees="default",
    weather="default",
    world_size="default",
    worms="default",
    task_set="cave_default",
    start_location="default",
    layout_mode="RestrictNodesByKey",
    roads="never",
    loop="default",
    branching="default",
    spawnmode="fixed",
    keep_disconnected_tiles=true,
    no_joining_islands=true,
    no_wormholes_to_disconnected_tiles=true,
    wormhole_prefab="tentacle_pillar"
  },
  required_prefabs={ "multiplayer_portal" },
  substitutes={  },
  version=4
}
`

// Keep old names for backward compatibility
const DefaultMasterWorldGen = DefaultMasterLevelData
const DefaultCavesWorldGen = DefaultCavesLevelData

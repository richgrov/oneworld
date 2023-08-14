package blocks

const Unbreakable float32 = -1
const InstaBreak float32 = 0

type blockProperties struct {
	hardness float32
}

var properties = [...]blockProperties{
	// Air
	{},
	// Stone
	{
		hardness: 1.5,
	},
	// Grass
	{
		hardness: 0.6,
	},
	// Dirt
	{
		hardness: 0.5,
	},
	// Cobblestone
	{
		hardness: 2.,
	},
	// Planks
	{
		hardness: 2.,
	},
	// Sapling
	{
		hardness: InstaBreak,
	},
	// Bedrock
	{
		hardness: Unbreakable,
	},
	// FlowingWater
	{
		hardness: 100,
	},
	// Water
	{
		hardness: 100,
	},
	// FlowingLava
	{
		hardness: InstaBreak,
	},
	// Lava
	{
		hardness: 100,
	},
	// Sand
	{
		hardness: 0.5,
	},
	// Gravel
	{
		hardness: 0.6,
	},
	// GoldOre
	{
		hardness: 3,
	},
	// IronOre
	{
		hardness: 3,
	},
	// CoalOre
	{
		hardness: 3,
	},
	// Log
	{
		hardness: 2,
	},
	// Leaves
	{
		hardness: 0.2,
	},
	// Sponge
	{
		hardness: 0.6,
	},
	// Glass
	{
		hardness: 0.3,
	},
	// LapisOre
	{
		hardness: 3,
	},
	// LapisBlock
	{
		hardness: 3,
	},
	// Dispenser
	{
		hardness: 3.5,
	},
	// Sandstone
	{
		hardness: 0.8,
	},
	// NoteBlock
	{
		hardness: 0.8,
	},
	// Bed
	{
		hardness: 0.2,
	},
	// PoweredRail
	{
		hardness: 0.7,
	},
	// DetectorRail
	{
		hardness: 0.7,
	},
	// StickyPiston
	{
		hardness: 0.5,
	},
	// Web
	{
		hardness: 4,
	},
	// TallGrass
	{
		hardness: InstaBreak,
	},
	// DeadBush
	{
		hardness: InstaBreak,
	},
	// Piston
	{
		hardness: 0.5,
	},
	// PistonHead
	{
		hardness: 0.5,
	},
	// Wool
	{
		hardness: 0.8,
	},
	// PistonExtension
	{
		hardness: Unbreakable,
	},
	// Dandelion
	{
		hardness: InstaBreak,
	},
	// Rose
	{
		hardness: InstaBreak,
	},
	// BrownMushroom
	{
		hardness: InstaBreak,
	},
	// RedMushroom
	{
		hardness: InstaBreak,
	},
	// GoldBlock
	{
		hardness: 3,
	},
	// IronBlock
	{
		hardness: 5,
	},
	// DoubleSlab
	{
		hardness: 2,
	},
	// Slab
	{
		hardness: 2,
	},
	// Bricks
	{
		hardness: 2,
	},
	// Tnt
	{
		hardness: InstaBreak,
	},
	// Bookshelf
	{
		hardness: 1.5,
	},
	// MossStone
	{
		hardness: 2,
	},
	// Obsidian
	{
		hardness: 10,
	},
	// Torch
	{
		hardness: InstaBreak,
	},
	// Fire
	{
		hardness: InstaBreak,
	},
	// Spawner
	{
		hardness: 5,
	},
	// WoodStairs
	{
		hardness: 2,
	},
	// Chest
	{
		hardness: 2.5,
	},
	// Redstone
	{
		hardness: InstaBreak,
	},
	// DiamondOre
	{
		hardness: 3,
	},
	// DiamondBlock
	{
		hardness: 5,
	},
	// CraftingTable
	{
		hardness: 2.5,
	},
	// Wheat
	{
		hardness: InstaBreak,
	},
	// Farmland
	{
		hardness: 0.6,
	},
	// Furnace
	{
		hardness: 3.5,
	},
	// LitFurnace
	{
		hardness: 3.5,
	},
	// StandingSign
	{
		hardness: 1,
	},
	// WoodenDoor
	{
		hardness: 3,
	},
	// Ladder
	{
		hardness: 0.4,
	},
	// Rail
	{
		hardness: 0.7,
	},
	// StoneStairs
	{
		hardness: 2,
	},
	// WallSign
	{
		hardness: 1,
	},
	// Lever
	{
		hardness: 0.5,
	},
	// StonePressurePlate
	{
		hardness: 0.5,
	},
	// IronDoor
	{
		hardness: 5,
	},
	// WoodPressurePlate
	{
		hardness: 0.5,
	},
	// RedstoneOre
	{
		hardness: 3,
	},
	// PoweredRedstoneOre
	{
		hardness: 3,
	},
	// RedstoneTorchOff
	{
		hardness: InstaBreak,
	},
	// RedstoneTorchOn
	{
		hardness: InstaBreak,
	},
	// StoneButton
	{
		hardness: 0.5,
	},
	// SnowLayer
	{
		hardness: 0.1,
	},
	// Ice
	{
		hardness: 0.5,
	},
	// Snow
	{
		hardness: 0.2,
	},
	// Cactus
	{
		hardness: 0.4,
	},
	// Clay
	{
		hardness: 0.6,
	},
	// SugarCane
	{
		hardness: InstaBreak,
	},
	// Jukebox
	{
		hardness: 2,
	},
	// Fence
	{
		hardness: 2,
	},
	// Pumpkin
	{
		hardness: 1,
	},
	// Netherrack
	{
		hardness: 0.4,
	},
	// SoulSand
	{
		hardness: 0.5,
	},
	// Glowstone
	{
		hardness: 0.3,
	},
	// Portal
	{
		hardness: Unbreakable,
	},
	// JackOLantern
	{
		hardness: 1,
	},
	// Cake
	{
		hardness: 0.5,
	},
	// RepeaterOff
	{
		hardness: InstaBreak,
	},
	// RepeaterOn
	{
		hardness: InstaBreak,
	},
	// LockedChest
	{
		hardness: InstaBreak,
	},
	// Trapdoor
	{
		hardness: 3,
	},
}

func Hardness(ty BlockType) float32 {
	return properties[ty].hardness
}

# Ravel Orchestrator Simulator 3D

> A Kerbal Space Program-style 3D visualization of the Ravel MicroVM orchestrator in action!

![Ravel Simulator](https://img.shields.io/badge/status-alpha-orange) ![Three.js](https://img.shields.io/badge/Three.js-v0.169-blue) ![TypeScript](https://img.shields.io/badge/TypeScript-5.6-blue)

## What is this?

This is an **interactive 3D game simulator** that visualizes how the Ravel orchestrator actually works. Walk around a virtual datacenter, watch machines spawn and transition through their lifecycle states, see resource allocation happen in real-time, and interact with the cluster - all while understanding the distributed systems magic happening behind the scenes.

Think **Kerbal Space Program**, but for container orchestration!

## Features

### Core Visualization
- **3D Cluster Topology**: Walk through regions containing multiple agent nodes
- **Machine Lifecycle FSM**: Watch machines transition through states with visual effects:
  - `created` → `preparing` → `starting` → `running` → `stopping` → `stopped` → `destroying` → `destroyed`
- **Resource Allocation**: Real-time visualization of CPU/Memory usage with particle effects
- **Event Stream**: Live feed of cluster events in the bottom-right HUD
- **Placement Algorithm**: Visual representation of agent selection and scoring

### Gameplay Controls (KSP-Style)
- **WASD** - Walk around the datacenter
- **Mouse** - Look around (FPS controls)
- **Space** - Jump
- **Shift** - Sprint
- **C** - Create a random machine (spawn workloads on the fly!)
- **Click** - Interact with machines and nodes
- **ESC** - Release mouse lock

### Visual Elements

#### Nodes (Agent Hosts)
- **Color-coded by status**:
  - 🟢 Green = Healthy (< 70% utilization)
  - 🟡 Yellow = Limited resources (70-90% utilization)
  - 🔴 Red = Exhausted (> 90% utilization)
- **Resource gauges**: Animated rings showing CPU and memory usage
- **Labels**: Display node name and region

#### Machines (MicroVMs)
- **State-based colors**: Each lifecycle state has a unique color
- **Rotation speed**: Indicates current state (faster = transitioning, slow = stable)
- **Particle effects**: Visual flourishes during state transitions
- **Connection lines**: Show which node hosts each machine

#### Environment
- **Grid floor**: Easy spatial reference
- **Ambient particles**: Atmospheric "datacenter vibes"
- **Fog**: Depth perception
- **Dynamic lighting**: Each node emits light based on its status

## Quick Start

### Prerequisites
- Node.js 18+
- npm or yarn
- A modern browser with WebGL support

### Installation

```bash
cd ravel-simulator-3d
npm install
```

### Running the Simulator

```bash
npm run dev
```

Or use npx:

```bash
npx vite
```

Then open your browser to the URL shown (typically `http://localhost:5173`).

Click anywhere in the window to lock your mouse and start exploring!

### Build for Production

```bash
npm run build
npm run preview
```

## Architecture

### Technology Stack
- **Three.js**: 3D rendering engine
- **TypeScript**: Type-safe development
- **Vite**: Lightning-fast dev server and build tool
- **GSAP**: Animation library (installed but ready for advanced animations)

### Project Structure

```
ravel-simulator-3d/
├── src/
│   ├── main.ts                 # Entry point
│   ├── types.ts                # Ravel API types (from actual Go codebase)
│   ├── controls/
│   │   └── FirstPersonControls.ts   # FPS camera controller
│   ├── entities/
│   │   ├── NodeEntity.ts       # 3D visualization of agent nodes
│   │   └── MachineEntity.ts    # 3D visualization of machines
│   ├── scene/
│   │   └── Scene.ts            # Main 3D scene management
│   └── simulator/
│       └── OrchestratorState.ts # Simulated Ravel orchestrator
├── index.html                   # Entry HTML with HUD
├── package.json
├── tsconfig.json
└── vite.config.ts
```

### How It Works

**OrchestratorState** simulates a real Ravel cluster:
1. Initializes 3 regions with multiple agent nodes
2. Tracks resources (CPU, memory, network interfaces)
3. Implements placement algorithm (finds best node for new machines)
4. Manages machine lifecycle with FSM-based state transitions
5. Emits events just like the real Ravel event stream
6. Handles resource allocation/deallocation

**Scene** renders everything:
1. Creates 3D entities for each node and machine
2. Updates entity states based on orchestrator state
3. Handles user interaction (clicking, creating machines)
4. Renders connection lines, particles, and effects

**FirstPersonControls** lets you walk around:
1. Pointer lock API for mouse look
2. WASD movement with physics
3. Gravity, jumping, and sprinting
4. Ground collision detection

## Understanding the Visualization

### What You're Seeing

- **Large cubes with glowing rings**: These are **agent nodes** (physical hosts running CloudHypervisor)
- **Small rotating cubes**: These are **machines** (individual microVMs running OCI images)
- **Lines connecting them**: Show which node is hosting each machine
- **Colored particles**: Resource flows and state transitions

### Machine Lifecycle

Watch a machine go through its full lifecycle:

1. **Created** (Blue) - Machine record created, waiting for placement
2. **Preparing** (Cyan, particles) - Pulling OCI image, setting up jailer sandbox
3. **Starting** (Light cyan, fast rotation) - CloudHypervisor starting the microVM
4. **Running** (Green, slow rotation) - Machine is live and serving traffic
5. **Stopping** (Orange) - Graceful shutdown initiated
6. **Stopped** (Dark orange, no rotation) - Machine halted but not destroyed
7. **Destroying** (Red, dissolving) - Cleaning up resources
8. **Destroyed** (Fades out) - Machine completely removed

### Resource Allocation

When you press **C** to create a machine:
1. Server broadcasts placement request to all agents in the region
2. Each agent checks if it has enough resources
3. Server selects the agent with the highest score (most available resources)
4. Machine spawns on the selected node
5. Resource gauges update to show new allocation

## Customization

### Creating Custom Scenarios

Edit `src/simulator/OrchestratorState.ts`:

```typescript
// Add more regions
const regionConfigs = [
  { id: 'ap-south', name: 'AP South', nodeCount: 5, basePos: { x: 40, y: 0, z: 0 } },
];

// Create custom machine templates
this.createMachine(
  'my-namespace',
  'my-fleet',
  'custom-image:latest',
  { cpuMhz: 2000, memoryMb: 4096, networkInterfaces: 2 }
);
```

### Adjusting Visual Style

Edit `index.html` for HUD styling or `src/scene/Scene.ts` for 3D rendering:

```typescript
// Change background color
this.scene.background = new THREE.Color(0x1a1a2e);

// Adjust fog
this.scene.fog = new THREE.Fog(0x1a1a2e, 50, 200);

// Change node colors in NodeEntity.ts
```

## Roadmap

### MVP Features (Current)
- [x] Basic 3D scene with first-person controls
- [x] Node and machine entities
- [x] Simulated orchestrator with placement algorithm
- [x] Machine lifecycle state machine
- [x] Resource allocation visualization
- [x] Event stream feed
- [x] Interactive machine creation

### Next Steps (Future)
- [ ] Namespace/Fleet hierarchy visualization (nested containers)
- [ ] Advanced particle systems for resource flows
- [ ] Gateway/Proxy traffic visualization (animated data packets)
- [ ] Network topology with IP allocation display
- [ ] Machine detail panels (click to inspect)
- [ ] Sound effects and atmospheric audio
- [ ] Performance metrics dashboard
- [ ] Real Ravel cluster connection mode (hybrid mode)
- [ ] Time controls (speed up/slow down simulation)
- [ ] Cluster management UI (start/stop/destroy machines via UI)

## Inspirations

- **Kerbal Space Program**: Complex systems made visual and interactive
- **Factorio**: Real-time resource flow visualization
- **Code Radio**: Combining work tools with ambient experiences
- **Ravel itself**: Making infrastructure beautiful and understandable

## Contributing

This is a fun experiment to make distributed systems education more engaging! Contributions welcome:

1. Fork the repo
2. Create a feature branch
3. Make your changes
4. Submit a PR

Ideas for contributions:
- New visual effects
- Additional simulation modes
- Performance optimizations
- Mobile/touch controls
- VR support (!!)

## License

Apache 2.0 (same as Ravel)

---

**Made with 🎮 by the Ravel community**

Shoutout to [Valyent](https://valyent.cloud) for creating Ravel!

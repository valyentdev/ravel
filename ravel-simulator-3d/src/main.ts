import { Scene } from './scene/Scene';
import { OrchestratorState } from './simulator/OrchestratorState';
import type { MachineEvent } from './types';

class RavelSimulator {
  private scene: Scene | null = null;
  private orchestrator: OrchestratorState;
  private animationFrameId: number | null = null;

  constructor() {
    this.orchestrator = new OrchestratorState();
    this.setupEventListeners();
    this.initialize();
  }

  private async initialize() {
    const loading = document.getElementById('loading');
    const container = document.getElementById('canvas-container');

    if (!container) {
      console.error('Canvas container not found');
      return;
    }

    // Simulate loading
    await new Promise(resolve => setTimeout(resolve, 1000));

    // Initialize scene
    this.scene = new Scene(container, this.orchestrator);

    // Hide loading screen
    if (loading) {
      loading.classList.add('hidden');
    }

    // Start render loop
    this.animate();

    // Start UI updates
    this.startUIUpdates();
  }

  private setupEventListeners() {
    // Listen to machine events
    this.orchestrator.onEvent((event: MachineEvent) => {
      this.addEventToFeed(event);
    });
  }

  private addEventToFeed(event: MachineEvent) {
    const eventList = document.getElementById('event-list');
    if (!eventList) return;

    const eventItem = document.createElement('div');
    eventItem.className = 'event-item';

    // Color based on event type
    if (event.type.includes('created') || event.type.includes('started')) {
      eventItem.classList.add('event-started');
    } else if (event.type.includes('stopped')) {
      eventItem.classList.add('event-stopped');
    } else if (event.type.includes('destroyed') || event.type.includes('failed')) {
      eventItem.classList.add('event-destroyed');
    } else {
      eventItem.classList.add('event-created');
    }

    const timestamp = new Date(event.timestamp).toLocaleTimeString();
    eventItem.innerHTML = `
      <div style="font-size: 12px; color: #888;">${timestamp}</div>
      <div>${event.message}</div>
    `;

    // Add to top of list
    eventList.insertBefore(eventItem, eventList.firstChild);

    // Keep only last 10 events
    while (eventList.children.length > 10) {
      eventList.removeChild(eventList.lastChild!);
    }
  }

  private startUIUpdates() {
    setInterval(() => {
      this.updateHUD();
    }, 100);
  }

  private updateHUD() {
    const stats = this.orchestrator.getClusterStats();

    // Update cluster stats
    const nodeCount = document.getElementById('node-count');
    const machineCount = document.getElementById('machine-count');
    const cpuUsage = document.getElementById('cpu-usage');
    const memoryUsage = document.getElementById('memory-usage');

    if (nodeCount) nodeCount.textContent = stats.nodeCount.toString();
    if (machineCount) machineCount.textContent = stats.machineCount.toString();
    if (cpuUsage) cpuUsage.textContent = `${stats.cpuUsage.toFixed(1)}%`;
    if (memoryUsage) memoryUsage.textContent = `${stats.memoryUsage.toFixed(1)}%`;

    // Update camera position
    if (this.scene) {
      const camera = this.scene.getCamera();
      const pos = camera.position;
      const cameraPos = document.getElementById('camera-pos');
      if (cameraPos) {
        cameraPos.textContent = `${pos.x.toFixed(1)}, ${pos.y.toFixed(1)}, ${pos.z.toFixed(1)}`;
      }
    }
  }

  private animate = () => {
    this.animationFrameId = requestAnimationFrame(this.animate);

    if (this.scene) {
      this.scene.update();
      this.scene.render();
    }
  };

  public dispose() {
    if (this.animationFrameId !== null) {
      cancelAnimationFrame(this.animationFrameId);
    }
  }
}

// Initialize the simulator
new RavelSimulator();

// Add some instructions to console
console.log(`
╔═══════════════════════════════════════════════════════╗
║   RAVEL ORCHESTRATOR SIMULATOR 3D                     ║
║   ─────────────────────────────────────────────────   ║
║   Controls:                                           ║
║     WASD - Move around                                ║
║     Mouse - Look around                               ║
║     Space - Jump                                      ║
║     Shift - Sprint                                    ║
║     C - Create random machine                         ║
║     Click - Interact with entities                    ║
║     ESC - Release mouse lock                          ║
║                                                       ║
║   Visualization:                                      ║
║     • Green nodes = Healthy                           ║
║     • Yellow nodes = Limited resources                ║
║     • Red nodes = Exhausted                           ║
║     • Machines change color by state                  ║
║     • Rings around nodes = Resource usage             ║
║                                                       ║
║   Have fun exploring the Ravel orchestrator!          ║
╚═══════════════════════════════════════════════════════╝
`);

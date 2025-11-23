import type {
  Node,
  Machine,
  MachineEvent,
  Region,
  Resources,
  MachineStatus,
  MachineEventType,
  PlacementRequest,
  PlacementResponse
} from '../types';

export class OrchestratorState {
  private regions: Map<string, Region> = new Map();
  private nodes: Map<string, Node> = new Map();
  private machines: Map<string, Machine> = new Map();
  private events: MachineEvent[] = [];
  private eventListeners: ((event: MachineEvent) => void)[] = [];

  constructor() {
    this.initializeCluster();
  }

  private initializeCluster() {
    // Create 3 regions with multiple nodes
    const regionConfigs = [
      { id: 'us-east', name: 'US East', nodeCount: 3, basePos: { x: -20, y: 0, z: 0 } },
      { id: 'us-west', name: 'US West', nodeCount: 3, basePos: { x: 0, y: 0, z: 0 } },
      { id: 'eu-central', name: 'EU Central', nodeCount: 2, basePos: { x: 20, y: 0, z: 0 } },
    ];

    regionConfigs.forEach(regionConfig => {
      const region: Region = {
        id: regionConfig.id,
        name: regionConfig.name,
        nodes: []
      };

      for (let i = 0; i < regionConfig.nodeCount; i++) {
        const node: Node = {
          id: `${regionConfig.id}-node-${i + 1}`,
          name: `node-${i + 1}`,
          region: regionConfig.id,
          totalResources: {
            cpuMhz: 4000 + Math.random() * 4000,
            memoryMb: 8192 + Math.random() * 8192,
            networkInterfaces: 4
          },
          availableResources: { cpuMhz: 0, memoryMb: 0, networkInterfaces: 0 },
          usedResources: { cpuMhz: 0, memoryMb: 0, networkInterfaces: 0 },
          status: 'healthy',
          position: {
            x: regionConfig.basePos.x + (i % 2) * 10 - 5,
            y: 0,
            z: regionConfig.basePos.z + Math.floor(i / 2) * 10
          }
        };

        node.availableResources = { ...node.totalResources };
        region.nodes.push(node);
        this.nodes.set(node.id, node);
      }

      this.regions.set(region.id, region);
    });

    // Create some initial machines for demo
    this.createRandomMachines(5);
  }

  private generateId(): string {
    return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }

  private createRandomMachines(count: number) {
    const namespaces = ['production', 'staging', 'development'];
    const fleets = ['web-servers', 'api-servers', 'workers', 'databases'];
    const images = ['nginx:latest', 'node:18-alpine', 'postgres:15', 'redis:7'];

    for (let i = 0; i < count; i++) {
      const namespace = namespaces[Math.floor(Math.random() * namespaces.length)];
      const fleet = fleets[Math.floor(Math.random() * fleets.length)];
      const image = images[Math.floor(Math.random() * images.length)];

      const resources: Resources = {
        cpuMhz: 1000 + Math.random() * 2000,
        memoryMb: 512 + Math.random() * 3584,
        networkInterfaces: 1
      };

      this.createMachine(namespace, fleet, image, resources);
    }
  }

  public createMachine(
    namespace: string,
    fleet: string,
    image: string,
    resources: Resources
  ): Machine | null {
    // Find best node using placement algorithm
    const placementRequest: PlacementRequest = {
      id: this.generateId(),
      region: 'us-east', // For simplicity, always use us-east
      resources,
      timestamp: Date.now()
    };

    const placement = this.findBestNode(placementRequest);
    if (!placement) {
      console.warn('No node available for placement');
      return null;
    }

    const node = this.nodes.get(placement.nodeId)!;
    const machine: Machine = {
      id: this.generateId(),
      name: `${fleet}-${Math.random().toString(36).substr(2, 6)}`,
      namespace,
      fleet,
      nodeId: node.id,
      status: 'created',
      image,
      resources,
      createdAt: Date.now(),
      position: {
        x: node.position.x + (Math.random() - 0.5) * 3,
        y: 2,
        z: node.position.z + (Math.random() - 0.5) * 3
      }
    };

    this.machines.set(machine.id, machine);
    this.allocateResources(node, resources);
    this.emitEvent(machine.id, 'machine_created', `Machine ${machine.name} created`);

    // Simulate state transitions
    setTimeout(() => this.transitionMachine(machine.id, 'preparing'), 500);

    return machine;
  }

  private findBestNode(request: PlacementRequest): PlacementResponse | null {
    let bestNode: Node | null = null;
    let bestScore = -1;

    for (const node of this.nodes.values()) {
      if (
        node.availableResources.cpuMhz >= request.resources.cpuMhz &&
        node.availableResources.memoryMb >= request.resources.memoryMb &&
        node.availableResources.networkInterfaces >= request.resources.networkInterfaces
      ) {
        // Score based on available resources
        const score =
          node.availableResources.cpuMhz +
          node.availableResources.memoryMb;

        if (score > bestScore) {
          bestScore = score;
          bestNode = node;
        }
      }
    }

    if (!bestNode) return null;

    return {
      nodeId: bestNode.id,
      score: bestScore
    };
  }

  private allocateResources(node: Node, resources: Resources) {
    node.usedResources.cpuMhz += resources.cpuMhz;
    node.usedResources.memoryMb += resources.memoryMb;
    node.usedResources.networkInterfaces += resources.networkInterfaces;

    node.availableResources.cpuMhz -= resources.cpuMhz;
    node.availableResources.memoryMb -= resources.memoryMb;
    node.availableResources.networkInterfaces -= resources.networkInterfaces;

    this.updateNodeStatus(node);
  }

  private deallocateResources(node: Node, resources: Resources) {
    node.usedResources.cpuMhz -= resources.cpuMhz;
    node.usedResources.memoryMb -= resources.memoryMb;
    node.usedResources.networkInterfaces -= resources.networkInterfaces;

    node.availableResources.cpuMhz += resources.cpuMhz;
    node.availableResources.memoryMb += resources.memoryMb;
    node.availableResources.networkInterfaces += resources.networkInterfaces;

    this.updateNodeStatus(node);
  }

  private updateNodeStatus(node: Node) {
    const cpuUsage = node.usedResources.cpuMhz / node.totalResources.cpuMhz;
    const memUsage = node.usedResources.memoryMb / node.totalResources.memoryMb;
    const usage = Math.max(cpuUsage, memUsage);

    if (usage >= 0.9) {
      node.status = 'exhausted';
    } else if (usage >= 0.7) {
      node.status = 'limited';
    } else {
      node.status = 'healthy';
    }
  }

  private transitionMachine(machineId: string, newStatus: MachineStatus) {
    const machine = this.machines.get(machineId);
    if (!machine) return;

    machine.status = newStatus;

    const eventTypeMap: Record<MachineStatus, MachineEventType | null> = {
      created: null,
      preparing: 'machine_prepare',
      starting: null,
      running: 'machine_started',
      stopping: null,
      stopped: 'machine_stopped',
      destroying: null,
      destroyed: 'machine_destroyed'
    };

    const eventType = eventTypeMap[newStatus];
    if (eventType) {
      this.emitEvent(machineId, eventType, `Machine ${machine.name} ${newStatus}`);
    }

    // Schedule next transition
    const transitions: Record<MachineStatus, [MachineStatus, number] | null> = {
      created: null,
      preparing: ['starting', 1000],
      starting: ['running', 1500],
      running: null,
      stopping: ['stopped', 1000],
      stopped: null,
      destroying: ['destroyed', 800],
      destroyed: null
    };

    const nextTransition = transitions[newStatus];
    if (nextTransition) {
      const [nextStatus, delay] = nextTransition;
      setTimeout(() => this.transitionMachine(machineId, nextStatus), delay);
    }
  }

  public startMachine(machineId: string) {
    const machine = this.machines.get(machineId);
    if (!machine || machine.status !== 'stopped') return;

    this.transitionMachine(machineId, 'preparing');
  }

  public stopMachine(machineId: string) {
    const machine = this.machines.get(machineId);
    if (!machine || machine.status !== 'running') return;

    this.transitionMachine(machineId, 'stopping');
  }

  public destroyMachine(machineId: string) {
    const machine = this.machines.get(machineId);
    if (!machine) return;

    const node = this.nodes.get(machine.nodeId);
    if (node) {
      this.deallocateResources(node, machine.resources);
    }

    this.transitionMachine(machineId, 'destroying');

    setTimeout(() => {
      this.machines.delete(machineId);
    }, 1000);
  }

  private emitEvent(machineId: string, type: MachineEventType, message: string) {
    const event: MachineEvent = {
      id: this.generateId(),
      machineId,
      type,
      timestamp: Date.now(),
      message
    };

    this.events.push(event);
    if (this.events.length > 100) {
      this.events.shift();
    }

    this.eventListeners.forEach(listener => listener(event));
  }

  public onEvent(listener: (event: MachineEvent) => void) {
    this.eventListeners.push(listener);
  }

  public getRegions(): Region[] {
    return Array.from(this.regions.values());
  }

  public getNodes(): Node[] {
    return Array.from(this.nodes.values());
  }

  public getMachines(): Machine[] {
    return Array.from(this.machines.values());
  }

  public getMachine(id: string): Machine | undefined {
    return this.machines.get(id);
  }

  public getNode(id: string): Node | undefined {
    return this.nodes.get(id);
  }

  public getRecentEvents(count: number = 20): MachineEvent[] {
    return this.events.slice(-count).reverse();
  }

  public getClusterStats() {
    let totalCpu = 0, usedCpu = 0, totalMem = 0, usedMem = 0;

    for (const node of this.nodes.values()) {
      totalCpu += node.totalResources.cpuMhz;
      usedCpu += node.usedResources.cpuMhz;
      totalMem += node.totalResources.memoryMb;
      usedMem += node.usedResources.memoryMb;
    }

    return {
      nodeCount: this.nodes.size,
      machineCount: this.machines.size,
      cpuUsage: totalCpu > 0 ? (usedCpu / totalCpu) * 100 : 0,
      memoryUsage: totalMem > 0 ? (usedMem / totalMem) * 100 : 0
    };
  }
}

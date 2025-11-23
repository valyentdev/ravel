// Ravel API Types based on the actual Go codebase

export type MachineStatus =
  | 'created'
  | 'preparing'
  | 'starting'
  | 'running'
  | 'stopping'
  | 'stopped'
  | 'destroying'
  | 'destroyed';

export type MachineEventType =
  | 'machine_created'
  | 'machine_prepare'
  | 'machine_started'
  | 'machine_stopped'
  | 'machine_destroyed'
  | 'machine_exited'
  | 'machine_failed';

export interface Resources {
  cpuMhz: number;
  memoryMb: number;
  networkInterfaces: number;
}

export interface Node {
  id: string;
  name: string;
  region: string;
  totalResources: Resources;
  availableResources: Resources;
  usedResources: Resources;
  status: 'healthy' | 'limited' | 'exhausted' | 'offline';
  position: { x: number; y: number; z: number };
}

export interface Machine {
  id: string;
  name: string;
  namespace: string;
  fleet: string;
  nodeId: string;
  status: MachineStatus;
  image: string;
  resources: Resources;
  createdAt: number;
  position: { x: number; y: number; z: number };
}

export interface MachineEvent {
  id: string;
  machineId: string;
  type: MachineEventType;
  timestamp: number;
  message: string;
}

export interface Region {
  id: string;
  name: string;
  nodes: Node[];
}

export interface Fleet {
  id: string;
  name: string;
  namespace: string;
  machines: Machine[];
}

export interface PlacementRequest {
  id: string;
  region: string;
  resources: Resources;
  timestamp: number;
}

export interface PlacementResponse {
  nodeId: string;
  score: number;
}

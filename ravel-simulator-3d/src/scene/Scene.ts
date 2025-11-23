import * as THREE from 'three';
import { FirstPersonControls } from '../controls/FirstPersonControls';
import { NodeEntity } from '../entities/NodeEntity';
import { MachineEntity } from '../entities/MachineEntity';
import { OrchestratorState } from '../simulator/OrchestratorState';
import { DatacenterEnvironment } from './DatacenterEnvironment';
import { CableSystem } from './CableSystem';
import type { Node, Machine } from '../types';

export class Scene {
  private scene: THREE.Scene;
  private camera: THREE.PerspectiveCamera;
  private renderer: THREE.WebGLRenderer;
  private controls: FirstPersonControls;
  private orchestrator: OrchestratorState;

  private nodeEntities: Map<string, NodeEntity> = new Map();
  private machineEntities: Map<string, MachineEntity> = new Map();

  private datacenterEnvironment: DatacenterEnvironment;
  private cableSystem: CableSystem;

  private raycaster: THREE.Raycaster = new THREE.Raycaster();
  private mouse: THREE.Vector2 = new THREE.Vector2();

  private clock: THREE.Clock = new THREE.Clock();

  constructor(container: HTMLElement, orchestrator: OrchestratorState) {
    this.orchestrator = orchestrator;

    // Setup scene
    this.scene = new THREE.Scene();
    this.scene.background = new THREE.Color(0x0a0a0f);
    this.scene.fog = new THREE.Fog(0x0a0a0f, 30, 150);

    // Setup camera
    this.camera = new THREE.PerspectiveCamera(
      75,
      window.innerWidth / window.innerHeight,
      0.1,
      1000
    );
    this.camera.position.set(0, 5, 15);

    // Setup renderer
    this.renderer = new THREE.WebGLRenderer({ antialias: true });
    this.renderer.setSize(window.innerWidth, window.innerHeight);
    this.renderer.setPixelRatio(window.devicePixelRatio);
    this.renderer.shadowMap.enabled = true;
    this.renderer.shadowMap.type = THREE.PCFSoftShadowMap;
    this.renderer.toneMapping = THREE.ACESFilmicToneMapping;
    this.renderer.toneMappingExposure = 1.2;
    container.appendChild(this.renderer.domElement);

    // Setup controls
    this.controls = new FirstPersonControls(this.camera, this.renderer.domElement);

    // Add lights
    this.setupLights();

    // Create datacenter environment
    this.datacenterEnvironment = new DatacenterEnvironment(this.orchestrator.getRegions());
    this.scene.add(this.datacenterEnvironment.getGroup());

    // Create cable system
    this.cableSystem = new CableSystem();
    this.scene.add(this.cableSystem.getGroup());

    // Add ambient particles
    this.createAmbientParticles();

    // Initialize entities
    this.initializeEntities();

    // Setup resize handler
    window.addEventListener('resize', () => this.onResize());

    // Setup click handler for interaction
    window.addEventListener('click', (event) => this.onClick(event));

    // Setup key handlers
    window.addEventListener('keydown', (event) => this.onKeyDown(event));
  }

  private setupLights() {
    // Ambient light
    const ambientLight = new THREE.AmbientLight(0x404040, 2);
    this.scene.add(ambientLight);

    // Directional light (sun)
    const directionalLight = new THREE.DirectionalLight(0xffffff, 1);
    directionalLight.position.set(50, 50, 50);
    directionalLight.castShadow = true;
    directionalLight.shadow.mapSize.width = 2048;
    directionalLight.shadow.mapSize.height = 2048;
    this.scene.add(directionalLight);

    // Hemisphere light for better ambient
    const hemisphereLight = new THREE.HemisphereLight(0x0080ff, 0x00ff80, 0.3);
    this.scene.add(hemisphereLight);
  }

  private createAmbientParticles() {
    const particlesGeometry = new THREE.BufferGeometry();
    const particlesCount = 1000;
    const positions = new Float32Array(particlesCount * 3);

    for (let i = 0; i < particlesCount * 3; i += 3) {
      positions[i] = (Math.random() - 0.5) * 200;
      positions[i + 1] = Math.random() * 50;
      positions[i + 2] = (Math.random() - 0.5) * 200;
    }

    particlesGeometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));

    const particlesMaterial = new THREE.PointsMaterial({
      color: 0x00ff00,
      size: 0.1,
      transparent: true,
      opacity: 0.3
    });

    const particles = new THREE.Points(particlesGeometry, particlesMaterial);
    this.scene.add(particles);
  }

  private initializeEntities() {
    // Create node entities
    for (const node of this.orchestrator.getNodes()) {
      const nodeEntity = new NodeEntity(node);
      this.nodeEntities.set(node.id, nodeEntity);
      this.scene.add(nodeEntity.getMesh());
    }

    // Create machine entities
    for (const machine of this.orchestrator.getMachines()) {
      const machineEntity = new MachineEntity(machine);
      this.machineEntities.set(machine.id, machineEntity);
      this.scene.add(machineEntity.getMesh());
    }

    // Initialize cables
    const nodesMap = new Map<string, Node>();
    for (const node of this.orchestrator.getNodes()) {
      nodesMap.set(node.id, node);
    }
    this.cableSystem.updateCables(this.orchestrator.getMachines(), nodesMap);
  }

  private onResize() {
    this.camera.aspect = window.innerWidth / window.innerHeight;
    this.camera.updateProjectionMatrix();
    this.renderer.setSize(window.innerWidth, window.innerHeight);
  }

  private onClick(event: MouseEvent) {
    if (!this.controls.getIsLocked()) return;

    // Raycast from camera center
    this.raycaster.setFromCamera(new THREE.Vector2(0, 0), this.camera);

    const allMeshes: THREE.Object3D[] = [];
    this.scene.traverse(object => {
      if (object.userData.type === 'node' || object.userData.type === 'machine') {
        allMeshes.push(object);
      }
    });

    const intersects = this.raycaster.intersectObjects(allMeshes, true);

    if (intersects.length > 0) {
      // Find the parent group
      let target = intersects[0].object;
      while (target.parent && !target.userData.type) {
        target = target.parent;
      }

      if (target.userData.type === 'machine') {
        const machine = target.userData.machine as Machine;
        console.log('Clicked machine:', machine);
        // TODO: Show machine details panel
      } else if (target.userData.type === 'node') {
        const node = target.userData.node as Node;
        console.log('Clicked node:', node);
        // TODO: Show node details panel
      }
    }
  }

  private onKeyDown(event: KeyboardEvent) {
    if (event.code === 'KeyC') {
      // Create a random machine
      this.orchestrator.createMachine(
        'production',
        'web-servers',
        'nginx:latest',
        { cpuMhz: 1000, memoryMb: 1024, networkInterfaces: 1 }
      );
    }
  }

  public update() {
    const delta = this.clock.getDelta();

    // Update controls
    this.controls.update(delta);

    // Update node entities
    for (const node of this.orchestrator.getNodes()) {
      const entity = this.nodeEntities.get(node.id);
      if (entity) {
        entity.update(node);
      }
    }

    // Update machine entities
    for (const machine of this.orchestrator.getMachines()) {
      let entity = this.machineEntities.get(machine.id);

      if (!entity) {
        // New machine, create entity
        entity = new MachineEntity(machine);
        this.machineEntities.set(machine.id, entity);
        this.scene.add(entity.getMesh());
      } else {
        entity.update(machine, delta);
      }
    }

    // Remove destroyed machines
    const currentMachineIds = new Set(
      this.orchestrator.getMachines().map(m => m.id)
    );

    for (const [id, entity] of this.machineEntities) {
      if (!currentMachineIds.has(id)) {
        this.scene.remove(entity.getMesh());
        this.machineEntities.delete(id);
      }
    }

    // Update cables
    const nodesMap = new Map<string, Node>();
    for (const node of this.orchestrator.getNodes()) {
      nodesMap.set(node.id, node);
    }
    this.cableSystem.updateCables(this.orchestrator.getMachines(), nodesMap);
    this.cableSystem.animateCables(delta);
  }

  public render() {
    this.renderer.render(this.scene, this.camera);
  }

  public getCamera(): THREE.PerspectiveCamera {
    return this.camera;
  }
}

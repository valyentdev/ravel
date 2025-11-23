import * as THREE from 'three';
import type { Node } from '../types';

export class NodeEntity {
  private node: Node;
  private mesh: THREE.Group;
  private resourceGauges: THREE.Group;
  private statusLight: THREE.PointLight;

  constructor(node: Node) {
    this.node = node;
    this.mesh = new THREE.Group();

    // Create server rack structure
    this.createServerRack();

    // Add status light on top
    this.statusLight = new THREE.PointLight(this.getStatusColor(), 3, 15);
    this.statusLight.position.set(0, 6, 0);
    this.mesh.add(this.statusLight);

    // Add resource gauges
    this.resourceGauges = this.createResourceGauges();
    this.mesh.add(this.resourceGauges);

    // Add label
    this.addLabel();

    // Position the mesh
    this.mesh.position.set(node.position.x, node.position.y, node.position.z);

    // Add hover effect (will be implemented with raycasting)
    this.mesh.userData = { type: 'node', node };
  }

  private createServerRack() {
    // Rack frame
    const frameGeometry = new THREE.BoxGeometry(3.5, 6, 2);
    const frameMaterial = new THREE.MeshStandardMaterial({
      color: 0x1a1a2e,
      roughness: 0.7,
      metalness: 0.5,
    });
    const frame = new THREE.Mesh(frameGeometry, frameMaterial);
    frame.position.y = 3;
    frame.castShadow = true;
    this.mesh.add(frame);

    // Rack door (front panel with vents)
    const doorGeometry = new THREE.BoxGeometry(3.4, 5.8, 0.1);
    const doorMaterial = new THREE.MeshStandardMaterial({
      color: 0x2a2a3a,
      roughness: 0.6,
      metalness: 0.6,
    });
    const door = new THREE.Mesh(doorGeometry, doorMaterial);
    door.position.set(0, 3, 1.05);
    door.castShadow = true;
    this.mesh.add(door);

    // Add server units (horizontal slots)
    for (let i = 0; i < 12; i++) {
      this.createServerUnit(i);
    }

    // Rack rails (vertical metal strips)
    const railGeometry = new THREE.BoxGeometry(0.1, 6, 0.1);
    const railMaterial = new THREE.MeshStandardMaterial({
      color: 0x666677,
      roughness: 0.8,
      metalness: 0.7,
    });

    const railPositions = [
      { x: -1.6, z: 0.95 },
      { x: 1.6, z: 0.95 },
      { x: -1.6, z: -0.95 },
      { x: 1.6, z: -0.95 },
    ];

    railPositions.forEach((pos) => {
      const rail = new THREE.Mesh(railGeometry, railMaterial);
      rail.position.set(pos.x, 3, pos.z);
      this.mesh.add(rail);
    });

    // Add cable management on the side
    this.addCableManagement();

    // Add status indicator panel
    this.addStatusPanel();
  }

  private createServerUnit(index: number) {
    const unitGroup = new THREE.Group();
    const yPos = -2.5 + index * 0.5;

    // Server unit body
    const unitGeometry = new THREE.BoxGeometry(3.2, 0.4, 1.8);
    const unitMaterial = new THREE.MeshStandardMaterial({
      color: 0x333344,
      roughness: 0.7,
      metalness: 0.5,
    });
    const unit = new THREE.Mesh(unitGeometry, unitMaterial);
    unit.position.set(0, yPos, 0);
    unitGroup.add(unit);

    // Front panel with LEDs
    const panelGeometry = new THREE.BoxGeometry(3.2, 0.35, 0.05);
    const panelMaterial = new THREE.MeshStandardMaterial({
      color: 0x222233,
      roughness: 0.6,
      metalness: 0.6,
    });
    const panel = new THREE.Mesh(panelGeometry, panelMaterial);
    panel.position.set(0, yPos, 0.95);
    unitGroup.add(panel);

    // Status LEDs
    const ledCount = 8;
    for (let i = 0; i < ledCount; i++) {
      const ledGeometry = new THREE.SphereGeometry(0.03, 8, 8);
      const isActive = Math.random() > 0.3;
      const ledColor = isActive ? this.getStatusColor() : 0x330000;
      const ledMaterial = new THREE.MeshBasicMaterial({
        color: ledColor,
        emissive: ledColor,
        emissiveIntensity: isActive ? 1 : 0.1,
      });
      const led = new THREE.Mesh(ledGeometry, ledMaterial);
      led.position.set(-1.3 + i * 0.3, yPos, 1);
      unitGroup.add(led);

      // Add small point light for active LEDs
      if (isActive) {
        const ledLight = new THREE.PointLight(ledColor, 0.2, 1);
        ledLight.position.set(-1.3 + i * 0.3, yPos, 1);
        unitGroup.add(ledLight);
      }
    }

    this.mesh.add(unitGroup);
  }

  private addCableManagement() {
    // Cable management arm on the side
    const armGeometry = new THREE.BoxGeometry(0.3, 6, 0.3);
    const armMaterial = new THREE.MeshStandardMaterial({
      color: 0x444455,
      roughness: 0.8,
      metalness: 0.4,
    });
    const arm = new THREE.Mesh(armGeometry, armMaterial);
    arm.position.set(2, 3, 0);
    this.mesh.add(arm);

    // Add some cable representations
    for (let i = 0; i < 6; i++) {
      const cableGeometry = new THREE.CylinderGeometry(0.05, 0.05, 1, 8);
      const cableColors = [0xff0000, 0x00ff00, 0x0000ff, 0xffff00, 0xff00ff, 0x00ffff];
      const cableMaterial = new THREE.MeshStandardMaterial({
        color: cableColors[i],
        roughness: 0.6,
      });
      const cable = new THREE.Mesh(cableGeometry, cableMaterial);
      cable.position.set(2, i + 0.5, 0);
      cable.rotation.z = Math.PI / 2;
      this.mesh.add(cable);
    }
  }

  private addStatusPanel() {
    // Top status panel
    const panelGeometry = new THREE.BoxGeometry(3, 0.4, 1.5);
    const panelMaterial = new THREE.MeshStandardMaterial({
      color: 0x1a1a2e,
      roughness: 0.5,
      metalness: 0.7,
      emissive: this.getStatusColor(),
      emissiveIntensity: 0.2,
    });
    const panel = new THREE.Mesh(panelGeometry, panelMaterial);
    panel.position.set(0, 6.2, 0);
    this.mesh.add(panel);

    // Status indicator
    const indicatorGeometry = new THREE.BoxGeometry(2.5, 0.3, 1);
    const indicatorMaterial = new THREE.MeshBasicMaterial({
      color: this.getStatusColor(),
      transparent: true,
      opacity: 0.7,
    });
    const indicator = new THREE.Mesh(indicatorGeometry, indicatorMaterial);
    indicator.position.set(0, 6.25, 0);
    this.mesh.add(indicator);
  }

  private getStatusColor(): number {
    switch (this.node.status) {
      case 'healthy':
        return 0x00ff00;
      case 'limited':
        return 0xffaa00;
      case 'exhausted':
        return 0xff0000;
      case 'offline':
        return 0x666666;
      default:
        return 0xffffff;
    }
  }

  private createResourceGauges(): THREE.Group {
    const gauges = new THREE.Group();
    gauges.position.set(-1.9, 3, 0);

    // CPU Gauge (vertical bar on side of rack)
    const cpuUsage = this.node.usedResources.cpuMhz / this.node.totalResources.cpuMhz;
    const cpuBar = this.createGaugeBar(cpuUsage, 0x00aaff, 'CPU');
    cpuBar.position.x = 0;
    gauges.add(cpuBar);

    // Memory Gauge (vertical bar on side of rack)
    const memUsage = this.node.usedResources.memoryMb / this.node.totalResources.memoryMb;
    const memBar = this.createGaugeBar(memUsage, 0xff00aa, 'MEM');
    memBar.position.x = -0.3;
    gauges.add(memBar);

    return gauges;
  }

  private createGaugeBar(usage: number, color: number, label: string): THREE.Group {
    const barGroup = new THREE.Group();
    const maxHeight = 5;
    const barHeight = usage * maxHeight;

    // Background bar
    const bgGeometry = new THREE.BoxGeometry(0.2, maxHeight, 0.5);
    const bgMaterial = new THREE.MeshStandardMaterial({
      color: 0x1a1a2e,
      roughness: 0.8,
      metalness: 0.3,
    });
    const bg = new THREE.Mesh(bgGeometry, bgMaterial);
    bg.position.y = 0;
    barGroup.add(bg);

    // Usage bar
    const usageGeometry = new THREE.BoxGeometry(0.18, barHeight, 0.48);
    const usageMaterial = new THREE.MeshStandardMaterial({
      color: color,
      emissive: color,
      emissiveIntensity: 0.5,
      roughness: 0.6,
      metalness: 0.4,
    });
    const usageBar = new THREE.Mesh(usageGeometry, usageMaterial);
    usageBar.position.y = -maxHeight / 2 + barHeight / 2;
    barGroup.add(usageBar);

    // Label
    const canvas = document.createElement('canvas');
    canvas.width = 128;
    canvas.height = 64;
    const ctx = canvas.getContext('2d')!;
    ctx.fillStyle = '#000000';
    ctx.fillRect(0, 0, 128, 64);
    ctx.fillStyle = `#${color.toString(16).padStart(6, '0')}`;
    ctx.font = 'bold 32px monospace';
    ctx.textAlign = 'center';
    ctx.fillText(label, 64, 45);

    const labelTexture = new THREE.CanvasTexture(canvas);
    const labelMaterial = new THREE.MeshBasicMaterial({ map: labelTexture, transparent: true });
    const labelGeometry = new THREE.PlaneGeometry(0.4, 0.2);
    const labelMesh = new THREE.Mesh(labelGeometry, labelMaterial);
    labelMesh.position.set(0, -maxHeight / 2 - 0.3, 0.26);
    barGroup.add(labelMesh);

    return barGroup;
  }

  private addLabel() {
    // Create a simple text sprite (we'll use canvas for this)
    const canvas = document.createElement('canvas');
    const context = canvas.getContext('2d')!;
    canvas.width = 256;
    canvas.height = 128;

    context.fillStyle = '#000000';
    context.fillRect(0, 0, 256, 128);

    context.fillStyle = '#00ff00';
    context.font = 'bold 24px monospace';
    context.textAlign = 'center';
    context.fillText(this.node.name, 128, 40);

    context.font = '16px monospace';
    context.fillText(`${this.node.region}`, 128, 70);

    const texture = new THREE.CanvasTexture(canvas);
    const spriteMaterial = new THREE.SpriteMaterial({ map: texture });
    const sprite = new THREE.Sprite(spriteMaterial);
    sprite.position.set(0, -4, 0);
    sprite.scale.set(4, 2, 1);

    this.mesh.add(sprite);
  }

  public update(node: Node) {
    this.node = node;

    // Update status light color
    const color = this.getStatusColor();
    this.statusLight.color.setHex(color);

    // Update status panel emissive color
    this.mesh.traverse((child) => {
      if (child instanceof THREE.Mesh) {
        const material = child.material as THREE.MeshStandardMaterial;
        if (material.emissive && material.emissive.getHex() !== 0) {
          // Update emissive colors
          if (child.position.y > 6) {
            // This is likely the status panel
            material.emissive.setHex(color);
          }
        }
      }
    });

    // Update resource gauges
    this.mesh.remove(this.resourceGauges);
    this.resourceGauges = this.createResourceGauges();
    this.mesh.add(this.resourceGauges);
  }

  public getMesh(): THREE.Group {
    return this.mesh;
  }

  public getNode(): Node {
    return this.node;
  }
}

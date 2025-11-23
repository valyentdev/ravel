import * as THREE from 'three';
import type { Machine, Node } from '../types';

interface Cable {
  id: string;
  machine: Machine;
  node: Node;
  tube: THREE.Mesh;
  particles: THREE.Points;
  particlePositions: Float32Array;
  particleProgress: number[];
}

export class CableSystem {
  private cables: Map<string, Cable> = new Map();
  private group: THREE.Group;

  constructor() {
    this.group = new THREE.Group();
  }

  public updateCables(machines: Machine[], nodes: Map<string, Node>) {
    const currentMachineIds = new Set(machines.map((m) => m.id));

    // Remove cables for destroyed machines
    for (const [id, cable] of this.cables) {
      if (!currentMachineIds.has(id)) {
        this.group.remove(cable.tube);
        this.group.remove(cable.particles);
        this.cables.delete(id);
      }
    }

    // Add or update cables for active machines
    for (const machine of machines) {
      const node = nodes.get(machine.nodeId);
      if (!node) continue;

      if (this.cables.has(machine.id)) {
        // Update existing cable
        this.updateCable(machine, node);
      } else {
        // Create new cable
        this.createCable(machine, node);
      }
    }
  }

  private createCable(machine: Machine, node: Node) {
    const startPos = new THREE.Vector3(
      machine.position.x,
      machine.position.y,
      machine.position.z
    );
    const endPos = new THREE.Vector3(
      node.position.x,
      node.position.y + 2,
      node.position.z
    );

    // Create cable path with slight curve
    const curve = this.createCableCurve(startPos, endPos);

    // Create cable tube
    const tube = this.createCableTube(curve, machine);

    // Create data flow particles
    const { particles, particlePositions, particleProgress } = this.createFlowParticles(curve, machine);

    const cable: Cable = {
      id: machine.id,
      machine,
      node,
      tube,
      particles,
      particlePositions,
      particleProgress,
    };

    this.cables.set(machine.id, cable);
    this.group.add(tube);
    this.group.add(particles);
  }

  private createCableCurve(start: THREE.Vector3, end: THREE.Vector3): THREE.CatmullRomCurve3 {
    // Calculate control points for a natural cable sag
    const midPoint = new THREE.Vector3(
      (start.x + end.x) / 2,
      Math.min(start.y, end.y) - 0.5, // Sag downward
      (start.z + end.z) / 2
    );

    // Add some randomness to make cables look more organic
    const offset1 = new THREE.Vector3(
      (Math.random() - 0.5) * 0.5,
      (Math.random() - 0.5) * 0.3,
      (Math.random() - 0.5) * 0.5
    );

    const offset2 = new THREE.Vector3(
      (Math.random() - 0.5) * 0.5,
      (Math.random() - 0.5) * 0.3,
      (Math.random() - 0.5) * 0.5
    );

    const point1 = new THREE.Vector3().lerpVectors(start, midPoint, 0.33).add(offset1);
    const point2 = new THREE.Vector3().lerpVectors(midPoint, end, 0.66).add(offset2);

    return new THREE.CatmullRomCurve3([start, point1, midPoint, point2, end]);
  }

  private createCableTube(curve: THREE.CatmullRomCurve3, machine: Machine): THREE.Mesh {
    const tubeGeometry = new THREE.TubeGeometry(curve, 32, 0.08, 8, false);

    // Color based on machine status
    const color = this.getCableColor(machine);
    const cableMaterial = new THREE.MeshStandardMaterial({
      color: color,
      emissive: color,
      emissiveIntensity: 0.3,
      roughness: 0.6,
      metalness: 0.4,
      transparent: true,
      opacity: 0.8,
    });

    const tube = new THREE.Mesh(tubeGeometry, cableMaterial);
    tube.castShadow = true;

    // Add outer glow
    const glowGeometry = new THREE.TubeGeometry(curve, 32, 0.12, 8, false);
    const glowMaterial = new THREE.MeshBasicMaterial({
      color: color,
      transparent: true,
      opacity: 0.2,
      side: THREE.BackSide,
    });
    const glow = new THREE.Mesh(glowGeometry, glowMaterial);
    tube.add(glow);

    return tube;
  }

  private createFlowParticles(curve: THREE.CatmullRomCurve3, machine: Machine): {
    particles: THREE.Points;
    particlePositions: Float32Array;
    particleProgress: number[];
  } {
    const particleCount = 20;
    const particlePositions = new Float32Array(particleCount * 3);
    const particleProgress: number[] = [];

    // Initialize particles along the curve
    for (let i = 0; i < particleCount; i++) {
      const t = i / particleCount;
      particleProgress.push(t);

      const point = curve.getPoint(t);
      particlePositions[i * 3] = point.x;
      particlePositions[i * 3 + 1] = point.y;
      particlePositions[i * 3 + 2] = point.z;
    }

    const particleGeometry = new THREE.BufferGeometry();
    particleGeometry.setAttribute('position', new THREE.BufferAttribute(particlePositions, 3));

    const color = this.getCableColor(machine);
    const particleMaterial = new THREE.PointsMaterial({
      color: color,
      size: 0.15,
      transparent: true,
      opacity: 0.8,
      blending: THREE.AdditiveBlending,
    });

    const particles = new THREE.Points(particleGeometry, particleMaterial);
    particles.userData = { curve, particleProgress };

    return { particles, particlePositions, particleProgress };
  }

  private getCableColor(machine: Machine): number {
    // Color based on machine status for visual feedback
    switch (machine.status) {
      case 'running':
        return 0x00ff88; // Green - healthy
      case 'starting':
      case 'preparing':
        return 0x00aaff; // Blue - starting up
      case 'stopping':
        return 0xffaa00; // Orange - stopping
      case 'stopped':
        return 0x666666; // Gray - stopped
      case 'destroying':
        return 0xff4400; // Red-orange - destroying
      default:
        return 0x00ddff; // Cyan - default
    }
  }

  private updateCable(machine: Machine, node: Node) {
    const cable = this.cables.get(machine.id);
    if (!cable) return;

    // Update cable color based on machine status
    const color = this.getCableColor(machine);
    const tubeMaterial = cable.tube.material as THREE.MeshStandardMaterial;
    tubeMaterial.color.setHex(color);
    tubeMaterial.emissive.setHex(color);

    const particleMaterial = cable.particles.material as THREE.PointsMaterial;
    particleMaterial.color.setHex(color);

    // Update glow
    if (cable.tube.children.length > 0) {
      const glow = cable.tube.children[0] as THREE.Mesh;
      const glowMaterial = glow.material as THREE.MeshBasicMaterial;
      glowMaterial.color.setHex(color);
    }

    cable.machine = machine;
    cable.node = node;
  }

  public animateCables(delta: number) {
    for (const cable of this.cables.values()) {
      // Only animate if machine is running or starting
      if (
        cable.machine.status !== 'running' &&
        cable.machine.status !== 'starting' &&
        cable.machine.status !== 'preparing'
      ) {
        continue;
      }

      const curve = cable.particles.userData.curve as THREE.CatmullRomCurve3;
      const particleProgress = cable.particles.userData.particleProgress as number[];
      const positions = cable.particlePositions;

      // Animate particles along the curve
      for (let i = 0; i < particleProgress.length; i++) {
        // Update progress
        particleProgress[i] += delta * 0.3; // Speed of data flow
        if (particleProgress[i] > 1) {
          particleProgress[i] = 0;
        }

        // Get point on curve
        const point = curve.getPoint(particleProgress[i]);
        positions[i * 3] = point.x;
        positions[i * 3 + 1] = point.y;
        positions[i * 3 + 2] = point.z;
      }

      // Update geometry
      const positionAttribute = cable.particles.geometry.getAttribute('position');
      positionAttribute.needsUpdate = true;

      // Pulse the cable glow for running machines
      if (cable.machine.status === 'running') {
        const tubeMaterial = cable.tube.material as THREE.MeshStandardMaterial;
        const pulse = Math.sin(Date.now() * 0.003) * 0.2 + 0.3;
        tubeMaterial.emissiveIntensity = pulse;
      }
    }
  }

  public getGroup(): THREE.Group {
    return this.group;
  }

  public getCables(): Map<string, Cable> {
    return this.cables;
  }
}

import * as THREE from 'three';
import type { Machine, MachineStatus } from '../types';

export class MachineEntity {
  private machine: Machine;
  private mesh: THREE.Group;
  private innerCube: THREE.Mesh;
  private rotationSpeed: number;

  constructor(machine: Machine) {
    this.machine = machine;
    this.mesh = new THREE.Group();
    this.rotationSpeed = 0;

    // Inner rotating cube
    const innerGeometry = new THREE.BoxGeometry(0.8, 0.8, 0.8);
    const innerMaterial = new THREE.MeshPhongMaterial({
      color: this.getStatusColor(),
      emissive: this.getStatusColor(),
      emissiveIntensity: 0.5,
      transparent: true,
      opacity: 0.9
    });

    this.innerCube = new THREE.Mesh(innerGeometry, innerMaterial);
    this.mesh.add(this.innerCube);

    // Outer wireframe
    const outerGeometry = new THREE.BoxGeometry(1, 1, 1);
    const edges = new THREE.EdgesGeometry(outerGeometry);
    const lineMaterial = new THREE.LineBasicMaterial({
      color: this.getStatusColor(),
      transparent: true,
      opacity: 0.6
    });
    const wireframe = new THREE.LineSegments(edges, lineMaterial);
    this.mesh.add(wireframe);

    // Particle effect for certain states
    if (this.machine.status === 'starting' || this.machine.status === 'preparing') {
      this.addParticles();
    }

    // Position the mesh
    this.mesh.position.set(machine.position.x, machine.position.y, machine.position.z);

    // Add to user data for raycasting
    this.mesh.userData = { type: 'machine', machine };

    // Set initial rotation speed
    this.updateRotationSpeed();
  }

  private getStatusColor(): number {
    const colors: Record<MachineStatus, number> = {
      created: 0x0088ff,
      preparing: 0x00aaff,
      starting: 0x00ddff,
      running: 0x00ff00,
      stopping: 0xffaa00,
      stopped: 0xff8800,
      destroying: 0xff4400,
      destroyed: 0xff0000
    };

    return colors[this.machine.status] || 0xffffff;
  }

  private updateRotationSpeed() {
    const speeds: Record<MachineStatus, number> = {
      created: 0.5,
      preparing: 2,
      starting: 3,
      running: 1,
      stopping: 0.5,
      stopped: 0,
      destroying: 4,
      destroyed: 0
    };

    this.rotationSpeed = speeds[this.machine.status] || 0;
  }

  private addParticles() {
    const particleCount = 20;
    const particles = new THREE.Group();

    for (let i = 0; i < particleCount; i++) {
      const geometry = new THREE.SphereGeometry(0.05, 8, 8);
      const material = new THREE.MeshBasicMaterial({
        color: this.getStatusColor(),
        transparent: true,
        opacity: 0.6
      });

      const particle = new THREE.Mesh(geometry, material);
      const angle = (i / particleCount) * Math.PI * 2;
      const radius = 1.5;

      particle.position.set(
        Math.cos(angle) * radius,
        Math.sin(i) * 0.5,
        Math.sin(angle) * radius
      );

      particle.userData = { angle, radius, offset: i };
      particles.add(particle);
    }

    particles.userData = { type: 'particles' };
    this.mesh.add(particles);
  }

  public update(machine: Machine, delta: number) {
    this.machine = machine;

    // Update color
    const color = this.getStatusColor();
    (this.innerCube.material as THREE.MeshPhongMaterial).color.setHex(color);
    (this.innerCube.material as THREE.MeshPhongMaterial).emissive.setHex(color);

    const wireframe = this.mesh.children[1] as THREE.LineSegments;
    (wireframe.material as THREE.LineBasicMaterial).color.setHex(color);

    // Update rotation
    this.updateRotationSpeed();
    this.innerCube.rotation.x += this.rotationSpeed * delta;
    this.innerCube.rotation.y += this.rotationSpeed * delta;

    // Update particles if they exist
    const particles = this.mesh.children.find(
      child => child.userData.type === 'particles'
    ) as THREE.Group | undefined;

    if (particles) {
      particles.children.forEach((particle, i) => {
        const { angle, radius, offset } = particle.userData;
        const time = Date.now() * 0.001 + offset;

        particle.position.x = Math.cos(angle + time) * radius;
        particle.position.y = Math.sin(time * 2) * 0.5;
        particle.position.z = Math.sin(angle + time) * radius;
      });

      // Remove particles if machine reaches stable state
      if (this.machine.status === 'running' || this.machine.status === 'stopped') {
        this.mesh.remove(particles);
      }
    }

    // Add dissolve effect for destroying state
    if (this.machine.status === 'destroying') {
      (this.innerCube.material as THREE.MeshPhongMaterial).opacity -= delta;
      (wireframe.material as THREE.LineBasicMaterial).opacity -= delta;
    }

    // Fade out for destroyed
    if (this.machine.status === 'destroyed') {
      (this.innerCube.material as THREE.MeshPhongMaterial).opacity = 0;
      (wireframe.material as THREE.LineBasicMaterial).opacity = 0;
    }
  }

  public getMesh(): THREE.Group {
    return this.mesh;
  }

  public getMachine(): Machine {
    return this.machine;
  }
}

<script lang="ts">
	import { onMount } from 'svelte';

	let email = $state('');
	let company = $state('');
	let currentDB = $state('');
	let submitted = $state(false);
	let submitting = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		submitting = true;
		
		// TODO: Replace with actual backend API call
		console.log('Waitlist signup:', { email, company, currentDB });
		
		// Simulate API call
		await new Promise(resolve => setTimeout(resolve, 1000));
		
		submitted = true;
		submitting = false;
	}

	onMount(() => {
		// Smooth scroll for anchor links
		document.querySelectorAll('a[href^="#"]').forEach(anchor => {
			anchor.addEventListener('click', function (e) {
				e.preventDefault();
				const target = document.querySelector(this.getAttribute('href') as string);
				if (target) {
					target.scrollIntoView({ behavior: 'smooth' });
				}
			});
		});
	});
</script>

<svelte:head>
	<title>VectorMigrate - Zero-Downtime Vector Database Migration</title>
	<meta name="description" content="Migrate between Pinecone, Weaviate, Qdrant, and Milvus with zero downtime. Automated schema translation, validation, and rollback." />
</svelte:head>

<div class="min-h-screen bg-gray-50">
	<!-- Navigation -->
	<nav class="bg-white shadow-sm sticky top-0 z-50">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
			<div class="flex justify-between items-center h-16">
				<div class="flex items-center gap-2">
					<span class="text-3xl">🔄</span>
					<span class="font-bold text-xl text-gray-900">VectorMigrate</span>
				</div>
				<div class="hidden md:flex items-center gap-8">
					<a href="#features" class="text-gray-600 hover:text-gray-900">Features</a>
					<a href="#how-it-works" class="text-gray-600 hover:text-gray-900">How It Works</a>
					<a href="#pricing" class="text-gray-600 hover:text-gray-900">Pricing</a>
					<a href="#faq" class="text-gray-600 hover:text-gray-900">FAQ</a>
					<button onclick={() => document.getElementById('waitlist')?.scrollIntoView({ behavior: 'smooth' })} class="btn-primary">
						Join Waitlist
					</button>
				</div>
			</div>
		</div>
	</nav>

	<!-- Hero Section -->
	<section class="gradient-bg hero-pattern text-white py-20 lg:py-32">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
			<h1 class="text-4xl md:text-6xl font-bold mb-6 leading-tight">
				Migrate Vector Databases<br/>
				<span class="text-indigo-200">Without the Headache</span>
			</h1>
			<p class="text-xl md:text-2xl text-indigo-100 mb-10 max-w-3xl mx-auto">
				Automated schema translation, zero-downtime migration, and validation between Pinecone, Weaviate, Qdrant, and Milvus.
				<br/><br/>
				<strong>Every week you're stuck in security review is a week your AI features aren't in production.</strong>
			</p>
			
			<!-- Waitlist Form -->
			<div id="waitlist" class="max-w-xl mx-auto bg-white rounded-2xl shadow-2xl p-8">
				{#if !submitted}
					<h3 class="text-2xl font-bold text-gray-900 mb-2">🚀 Get Early Access</h3>
					<p class="text-gray-600 mb-6">Join 500+ engineers on the waitlist. Launching Q2 2026.</p>
					
					<form onsubmit={handleSubmit} class="space-y-4">
						<div>
							<input 
								type="email" 
								bind:value={email}
								placeholder="you@company.com" 
								required
								class="input"
							/>
						</div>
						<div>
							<input 
								type="text" 
								bind:value={company}
								placeholder="Company name" 
								required
								class="input"
							/>
						</div>
						<div>
							<select 
								bind:value={currentDB}
								required
								class="input bg-white"
							>
								<option value="" disabled>Current vector DB?</option>
								<option value="pinecone">Pinecone</option>
								<option value="weaviate">Weaviate</option>
								<option value="qdrant">Qdrant</option>
								<option value="milvus">Milvus</option>
								<option value="other">Other / Evaluating</option>
							</select>
						</div>
						<button 
							type="submit" 
							disabled={submitting}
							class="w-full btn-primary text-lg py-4"
						>
							{#if submitting}
								Joining...
							{:else}
								Join Waitlist →
							{/if}
						</button>
					</form>
					
					<p class="text-xs text-gray-500 mt-4">
						🔒 We respect your inbox. No spam, ever. Unsubscribe anytime.
					</p>
				{:else}
					<div class="text-center py-8">
						<div class="text-6xl mb-4">✅</div>
						<h3 class="text-2xl font-bold text-green-800 mb-2">You're on the list!</h3>
						<p class="text-green-700">We'll send you early access details soon.</p>
					</div>
				{/if}
			</div>
			
			<!-- Social Proof -->
			<div class="mt-12 text-indigo-200">
				<p class="text-sm mb-4">Trusted by engineers from</p>
				<div class="flex justify-center items-center gap-8 opacity-70 flex-wrap">
					<span class="text-lg font-semibold">Stripe</span>
					<span class="text-lg font-semibold">Notion</span>
					<span class="text-lg font-semibold">Perplexity</span>
					<span class="text-lg font-semibold">Cohere</span>
				</div>
			</div>
		</div>
	</section>

	<!-- Problem Section -->
	<section class="py-20 bg-white">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
			<div class="text-center mb-16">
				<h2 class="text-3xl md:text-4xl font-bold text-gray-900 mb-4">
					Why Migrate Vector Databases?
				</h2>
				<p class="text-xl text-gray-600 max-w-3xl mx-auto">
					Companies outgrow their vector DB, hit security blockers, or need better pricing. But migration is painful.
				</p>
			</div>
			
			<div class="grid md:grid-cols-3 gap-8">
				<div class="bg-red-50 rounded-2xl p-8 border border-red-100">
					<div class="text-4xl mb-4">🔒</div>
					<h3 class="text-xl font-bold text-gray-900 mb-3">Security Review Blockers</h3>
					<p class="text-gray-700">
						"Every week you're stuck in security review is a week your AI features aren't in production." 
						— Pinecone BYOC announcement, Feb 2026
					</p>
				</div>
				
				<div class="bg-orange-50 rounded-2xl p-8 border border-orange-100">
					<div class="text-4xl mb-4">💸</div>
					<h3 class="text-xl font-bold text-gray-900 mb-3">Cost Explosion</h3>
					<p class="text-gray-700">
						Vector DB bills grow 10x as you scale. Companies need to migrate to self-hosted or cheaper alternatives, but can't afford downtime.
					</p>
				</div>
				
				<div class="bg-yellow-50 rounded-2xl p-8 border border-yellow-100">
					<div class="text-4xl mb-4">🔧</div>
					<h3 class="text-xl font-bold text-gray-900 mb-3">Manual Migration Hell</h3>
					<p class="text-gray-700">
						Schema differences, embedding validation, index optimization—each requires custom scripts. One mistake = corrupted vectors.
					</p>
				</div>
			</div>
		</div>
	</section>

	<!-- Features Section -->
	<section id="features" class="py-20 bg-gray-50">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
			<div class="text-center mb-16">
				<h2 class="text-3xl md:text-4xl font-bold text-gray-900 mb-4">
					Everything You Need for Migration
				</h2>
				<p class="text-xl text-gray-600 max-w-3xl mx-auto">
					Automated tools to move millions of vectors between databases without losing sleep.
				</p>
			</div>
			
			<div class="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
				<div class="card">
					<div class="text-3xl mb-4">🔄</div>
					<h3 class="text-xl font-bold text-gray-900 mb-3">Schema Auto-Translation</h3>
					<p class="text-gray-700">
						Automatically maps metadata, indexes, and configurations between Pinecone ↔ Weaviate ↔ Qdrant ↔ Milvus.
					</p>
				</div>
				
				<div class="card">
					<div class="text-3xl mb-4">⚡</div>
					<h3 class="text-xl font-bold text-gray-900 mb-3">Zero-Downtime Sync</h3>
					<p class="text-gray-700">
						Dual-write during migration. Switch traffic with confidence. Rollback in seconds if needed.
					</p>
				</div>
				
				<div class="card">
					<div class="text-3xl mb-4">✅</div>
					<h3 class="text-xl font-bold text-gray-900 mb-3">Embedding Validation</h3>
					<p class="text-gray-700">
						Cosine similarity checks ensure vectors survive migration intact. 99.9% preservation guarantee.
					</p>
				</div>
				
				<div class="card">
					<div class="text-3xl mb-4">📊</div>
					<h3 class="text-xl font-bold text-gray-900 mb-3">Performance Benchmarks</h3>
					<p class="text-gray-700">
						Before/after latency, recall, and throughput reports. Prove migration success to stakeholders.
					</p>
				</div>
				
				<div class="card">
					<div class="text-3xl mb-4">🔐</div>
					<h3 class="text-xl font-bold text-gray-900 mb-3">Security Compliance</h3>
					<p class="text-gray-700">
						SOC 2, HIPAA, GDPR-ready migration paths. Audit logs, encryption in transit/at rest.
					</p>
				</div>
				
				<div class="card">
					<div class="text-3xl mb-4">🚀</div>
					<h3 class="text-xl font-bold text-gray-900 mb-3">CLI + API</h3>
					<p class="text-gray-700">
						Command-line for engineers, REST API for automation. Integrate with CI/CD pipelines.
					</p>
				</div>
			</div>
		</div>
	</section>

	<!-- Footer -->
	<footer class="bg-gray-900 text-gray-300 py-12">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
			<div class="flex items-center justify-center gap-2 mb-4">
				<span class="text-2xl">🔄</span>
				<span class="font-bold text-white text-lg">VectorMigrate</span>
			</div>
			<p class="text-sm">&copy; 2026 VectorMigrate. Built with ❤️ using SvelteKit + Go.</p>
			<div class="mt-4 space-x-4 text-sm">
				<a href="https://github.com/AlphaTechini/vector-db-migration" class="hover:text-white">GitHub</a>
				<a href="#" class="hover:text-white">Privacy</a>
				<a href="#" class="hover:text-white">Terms</a>
			</div>
		</div>
	</footer>
</div>

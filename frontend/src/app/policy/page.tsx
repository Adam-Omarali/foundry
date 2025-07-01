export default function PrivacyPolicy() {
  return (
    <div className="max-w-3xl mx-auto py-12 px-4">
      <h1 className="text-3xl font-bold mb-8">Privacy Policy</h1>

      <section className="mb-8">
        <h2 className="text-xl font-semibold mb-4">Introduction</h2>
        <p className="mb-4">
          This privacy policy describes how Foundry ("we", "us", or "our")
          collects, uses, and shares data when you use our Chrome extension and
          associated web application. By using Foundry, you agree to the
          collection and use of information in accordance with this policy.
        </p>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold mb-4">Information Collection</h2>
        <p className="mb-4">We collect the following types of information:</p>
        <ul className="list-disc ml-6 mb-4">
          <li>
            Authentication data through Google Sign-in (email address and
            profile information). This is only used to authenticate you and only
            your email address is stored on our servers.
          </li>
          <li>Content from web pages you explicitly choose to save</li>
          <li>Extension usage data (only the number of pages saved)</li>
        </ul>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold mb-4">Data Usage</h2>
        <p className="mb-4">Your data is used for the following purposes:</p>
        <ul className="list-disc ml-6 mb-4">
          <li>Account authentication and management</li>
          <li>Storing and retrieving your saved web pages</li>
          <li>Generating embeddings for semantic search functionality</li>
          <li>Analyzing usage patterns to improve our service</li>
          <li>Debugging and maintaining service quality</li>
        </ul>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold mb-4">Data Sharing</h2>
        <p className="mb-4">
          We share your data with the following third parties:
        </p>
        <ul className="list-disc ml-6 mb-4">
          <li>Google Authentication - for user authentication services</li>
          <li>
            Google Gemini API - for generating embeddings and processing content
          </li>
        </ul>
        <p className="mb-4">
          These service providers are contractually bound to handle your data
          securely and cannot use it for their own purposes. We do not sell your
          personal information to any third parties.
        </p>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold mb-4">
          Data Storage and Security
        </h2>
        <p className="mb-4">
          All data is stored in a Supabase database located in the United
          States. We implement appropriate security measures including:
        </p>
        <ul className="list-disc ml-6 mb-4">
          <li>Secure storage via authentication tokens and JWT</li>
          <li>Regular security audits and updates</li>
          <li>Access controls and monitoring</li>
        </ul>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold mb-4">User Rights and Control</h2>
        <p className="mb-4">
          You have the following rights regarding your data:
        </p>
        <ul className="list-disc ml-6 mb-4">
          <li>Access your stored data through the web interface</li>
          <li>Delete individual saved pages</li>
          <li>Delete your entire account and associated data</li>
          <li>Opt-out of non-essential data collection</li>
        </ul>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold mb-4">
          Updates to Privacy Policy
        </h2>
        <p className="mb-4">
          We may update this privacy policy from time to time. We will notify
          you of any changes by posting the new policy on this page and updating
          the "last modified" date. Continued use of Foundry after any
          modifications indicates your acceptance of the updated policy.
        </p>
        <p className="mb-4">Last modified: March 14, 2024</p>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold mb-4">
          Chrome Web Store Compliance
        </h2>
        <p className="mb-4">
          This privacy policy complies with Chrome Web Store requirements and
          can be accessed through our Chrome Web Store listing. Our extension
          requests only the minimum permissions necessary to function, and we
          maintain transparency about all data collection and usage.
        </p>
      </section>

      <section>
        <h2 className="text-xl font-semibold mb-4">Contact Information</h2>
        <p className="mb-4">
          For any questions or concerns about this privacy policy or our data
          practices, please contact us:
        </p>
        <ul className="list-none mb-4">
          <li>Email: adamomarali37@gmail.com</li>
        </ul>
      </section>
    </div>
  );
}

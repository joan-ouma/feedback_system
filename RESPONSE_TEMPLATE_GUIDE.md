# Professional Response Template Guide

## Overview

A polished, professional response template has been added to display structured feedback responses with empathy, crisis resources, and actionable tips.

## Files Created/Modified

### 1. Template File
- **Location:** `templates/response.html`
- **Features:**
  - Tailwind CSS for modern, responsive design
  - Font Awesome icons for visual appeal
  - Glass morphism effects
  - Mobile-first responsive layout
  - Crisis resources section (shown when needed)
  - Actionable tips grid
  - Professional color palette

### 2. Handler Updates
- **File:** `internal/handlers/consultation_handler.go`
- **Added:**
  - `ResponseData` struct for structured template data
  - `Tip` struct for individual tips
  - `ShowResponse()` handler method
  - `GenerateResponseData()` helper function

## Usage

### Route
The template is accessible at: `/consultation/response`

### Example Usage in Handler

```go
// Simple example - render with default data
data := ResponseData{
    EmpathyTitle: "I'm truly sorry you're feeling this way",
    EmpathyMessage: "That sounds incredibly difficult...",
    ShowCrisisResources: true,
    Tips: []Tip{
        {
            Title: "Break things down",
            Description: "When overwhelmed, split tasks...",
            Icon: "fas fa-puzzle-piece",
            IconColor: "text-green-500",
            ColorClass: "border-green-400",
        },
        // ... more tips
    },
}
```

### Using GenerateResponseData Helper

```go
// Generate response data from LLM response
responseData := GenerateResponseData(llmResponse, userMessage)

// This automatically:
// - Detects crisis keywords
// - Sets appropriate empathy messages
// - Provides default helpful tips
```

## Integration with Existing System

### Option 1: Standalone Response Page
Users can visit `/consultation/response` to see a professional response page.

### Option 2: Integrate with Consultation Flow
Modify the consultation handler to redirect to this page after certain responses:

```go
// In SendMessage handler, after getting LLM response:
if shouldShowProfessionalResponse(consultation.Response) {
    // Redirect to professional response page
    http.Redirect(w, r, "/consultation/response", http.StatusSeeOther)
    return
}
```

### Option 3: Use as Modal/Component
The template can be adapted to be used as a modal or component within the consultation chat interface.

## Customization

### Tips
Tips can be customized based on:
- User's specific concerns (detected from message)
- LLM response analysis
- User's mood/assessment scores
- Time of day or context

### Crisis Detection
The `GenerateResponseData` function automatically detects crisis keywords:
- "suicide"
- "hurt myself"
- "end it all"
- "not worth living"

Add more keywords as needed in the function.

### Styling
The template uses Tailwind CSS classes. Customize:
- Colors: Change `text-{color}-500` classes
- Borders: Modify `border-{color}-400` classes
- Icons: Replace Font Awesome icon classes
- Layout: Adjust grid columns and spacing

## Features

✅ **Professional Design**
- Glass morphism effects
- Smooth animations
- Hover effects
- Responsive grid

✅ **Crisis Support**
- Prominent crisis resources
- Emergency contact numbers
- 24/7 support information

✅ **Actionable Tips**
- Visual icons
- Color-coded categories
- Hover animations
- Mobile-responsive grid

✅ **Empathy Section**
- Warm, supportive messaging
- Visual icon support
- Gradient backgrounds

## Testing

1. **Access the route:**
   ```
   GET /consultation/response
   ```
   (Requires authentication)

2. **Verify:**
   - Template loads correctly
   - All tips display
   - Crisis resources show when `ShowCrisisResources: true`
   - Responsive design works on mobile
   - Icons and styling render properly

## Next Steps

1. **Enhance LLM Integration:**
   - Analyze LLM responses to generate contextual tips
   - Detect user sentiment to customize empathy messages
   - Extract specific concerns to provide targeted advice

2. **Add More Tips:**
   - Create a tips database
   - Categorize tips by concern type
   - Personalize tips based on user history

3. **A/B Testing:**
   - Test different empathy messages
   - Compare tip effectiveness
   - Optimize crisis resource placement

## Browser Support

- Modern browsers (Chrome, Firefox, Safari, Edge)
- Mobile browsers (iOS Safari, Chrome Mobile)
- Requires JavaScript for Tailwind CSS (CDN)

## Dependencies

- Tailwind CSS 2.2.19 (CDN)
- Font Awesome 6.0.0 (CDN)

These are loaded via CDN, so no local installation needed.

